// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"

	null "github.com/guregu/null/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/migrations"
	"github.com/ubccr/grendel/internal/store/sqlstore/db"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
	"golang.org/x/crypto/bcrypt"
)

// SqlStore implements a Grendel Store using sqlc
type SqlStore struct {
	q  *db.Queries
	rw *sql.DB
	ro *sql.DB
}

// New returns a new SqlStore using the given database filename. For memory only you can provide `:memory:`
func New(filename string, config ...Config) (*SqlStore, error) {
	cfg := configDefault(config...)

	rw, err := sql.Open(cfg.Driver, cfg.DataSourceName(filename, true))
	if err != nil {
		return nil, err
	}

	migrator, err := migrations.New(rw)
	if err != nil {
		return nil, err
	}

	cur, dirty, err := migrator.Version()
	if err != nil && err != migrations.ErrNilVersion {
		return nil, fmt.Errorf("Failed to get current db version: %w", err)
	}

	store.Log.WithFields(logrus.Fields{
		"version": cur,
		"dirty":   dirty,
	}).Info("Current version")

	err = migrator.Migrate()
	if err != nil && err != migrations.ErrNoChange {
		return nil, err
	}

	if err == migrations.ErrNoChange {
		store.Log.Info("Database up to date, no new migrations")
	} else {
		store.Log.WithFields(logrus.Fields{
			"version": migrations.SchemaVersion,
		}).Info("Database migrated")
	}

	var ro *sql.DB
	if filename != ":memory:" {
		var err error
		ro, err = sql.Open(cfg.Driver, cfg.DataSourceName(filename, false))
		if err != nil {
			return nil, err
		}
	} else {
		ro = rw
	}

	rw.SetMaxOpenConns(1)

	return &SqlStore{rw: rw, ro: ro, q: db.New()}, nil
}

// StoreUser stores the User in the data store
func (s *SqlStore) StoreUser(username, password string) (string, error) {
	ctx := context.Background()

	role := model.RoleUser
	enabled := false

	count, err := s.q.UserCount(ctx, s.ro)
	if err != nil {
		return "", err
	}

	if count == 0 {
		role = model.RoleAdmin
		enabled = true
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", err
	}

	_, err = s.q.UserCreate(ctx, s.rw, db.UserCreateParams{
		Username:     username,
		Role:         role.String(),
		PasswordHash: string(hashed),
		Enabled:      enabled,
	})
	if err != nil {
		return "", err
	}

	return role.String(), nil
}

// VerifyUser checks if the given username exists in the data store
func (s *SqlStore) VerifyUser(username, password string) (bool, string, error) {
	ctx := context.Background()

	user, err := s.q.UserFetch(ctx, s.ro, username)
	if err != nil {
		return false, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, "", err
	}

	return true, user.Role, nil
}

// GetUsers returns a list of all the usernames
func (s *SqlStore) GetUsers() ([]model.User, error) {
	ctx := context.Background()

	userList, err := s.q.UserList(ctx, s.ro)
	if err != nil {
		return nil, err
	}

	users := make([]model.User, len(userList))
	for i, u := range userList {
		users[i] = model.User{
			ID:           u.ID,
			Username:     u.Username,
			Role:         u.Role,
			Enabled:      u.Enabled,
			PasswordHash: u.PasswordHash,
			CreatedAt:    u.CreatedAt,
			ModifiedAt:   u.UpdatedAt,
		}
	}

	return users, nil
}

// GetUsers returns a list of all the usernames
func (s *SqlStore) GetUserByName(name string) (*model.User, error) {
	ctx := context.Background()

	user, err := s.q.UserFetch(ctx, s.ro, name)
	if err != nil {
		return nil, err
	}

	output := model.User{
		ID:           user.ID,
		Username:     user.Username,
		Role:         user.Role,
		Enabled:      user.Enabled,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		ModifiedAt:   user.UpdatedAt,
	}
	return &output, nil
}

// UpdateUser updates the role of the given users
func (s *SqlStore) UpdateUserRole(username, role string) error {
	ctx := context.Background()

	err := s.q.UserUpdateRole(ctx, s.rw, db.UserUpdateRoleParams{
		Username: username,
		Role:     role,
	})

	return err
}

// UpdateUser updates the role of the given users
func (s *SqlStore) UpdateUserEnabled(username string, enabled bool) error {
	ctx := context.Background()

	err := s.q.UserUpdateEnable(ctx, s.rw, db.UserUpdateEnableParams{
		Username: username,
		Enabled:  enabled,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes the given user
func (s *SqlStore) DeleteUser(username string) error {
	ctx := context.Background()

	err := s.q.UserDelete(ctx, s.rw, username)
	if err != nil {
		return err
	}
	return nil
}

// StoreHost stores a host in the data store. If the host exists it is overwritten
func (s *SqlStore) StoreHost(host *model.Host) error {
	return s.StoreHosts(model.HostList{host})
}

// StoreHosts stores a list of host in the data store. If the host exists it is overwritten
func (s *SqlStore) StoreHosts(hosts model.HostList) error {
	ctx := context.Background()
	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for idx, h := range hosts {
		if h.Name == "" {
			return fmt.Errorf("host name required for host %d: %w", idx, store.ErrInvalidData)
		}

		// Link kernel if exists
		kernelID := null.NewInt(0, false)
		kernel, err := s.q.KernelFetch(ctx, tx, h.BootImage)
		if err == nil {
			kernelID.SetValid(kernel.ID)
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if h.UID.IsNil() {
			h.UID, err = ksuid.NewRandom()
			if err != nil {
				return err
			}
		}

		// Upsert node
		node, err := s.q.NodeUpsert(ctx, tx, db.NodeUpsertParams{
			ID:        null.NewInt(h.ID, h.ID != 0),
			UID:       h.UID,
			KernelID:  kernelID,
			Name:      h.Name,
			Provision: h.Provision,
			Firmware:  null.NewString(h.Firmware.String(), !h.Firmware.IsNil()),
		})
		if err != nil {
			return err
		}

		// Upsert tags
		tagIDs := make([]int64, 0)
		for _, t := range h.Tags {
			tg, err := s.q.TagUpsert(ctx, tx, t)
			if err != nil {
				return err
			}

			err = s.q.NodeTagUpsert(ctx, tx, db.NodeTagUpsertParams{
				TagID:  tg.ID,
				NodeID: node.ID,
				Value:  "", // TODO support key-value pairs
			})
			if err != nil {
				return err
			}
			tagIDs = append(tagIDs, tg.ID)
		}

		// Delete any tags that were removed
		err = s.q.NodeTagUpsertDelete(ctx, tx, db.NodeTagUpsertDeleteParams{
			Nodes: []int64{node.ID},
			Tags:  tagIDs,
		})
		if err != nil {
			return err
		}

		// Upsert network interfaces
		nicIDs := make([]int64, 0)
		for _, n := range h.Interfaces {
			nt := model.NicTypeEthernet
			if n.BMC {
				nt = model.NicTypeBMC
			}
			nc, err := s.q.NicUpsert(ctx, tx, db.NicUpsertParams{
				ID:      null.NewInt(n.ID, n.ID != 0),
				NodeID:  node.ID,
				NicType: nt.String(),
				Name:    null.NewString(n.Name, len(n.Name) != 0),
				IP:      null.NewString(n.IP.String(), n.IP.IsValid()),
				MAC:     null.NewString(n.MAC.String(), n.MAC != nil),
				FQDN:    null.NewString(n.FQDN, len(n.FQDN) != 0),
				VLAN:    null.NewString(n.VLAN, len(n.VLAN) != 0),
				MTU:     null.NewInt(int64(n.MTU), n.MTU != 0),
			})
			if err != nil {
				return err
			}
			nicIDs = append(nicIDs, nc.ID)
		}

		// Upsert bond interfaces
		for _, n := range h.Bonds {
			peers := null.NewString("", false)
			if len(n.Peers) > 0 {
				pj, err := json.Marshal(&n.Peers)
				if err != nil {
					return err
				}
				peers.SetValid(string(pj))
			}
			bi, err := s.q.NicUpsert(ctx, tx, db.NicUpsertParams{
				ID:      null.NewInt(n.ID, n.ID != 0),
				NodeID:  node.ID,
				NicType: model.NicTypeBond.String(),
				Name:    null.NewString(n.Name, len(n.Name) != 0),
				IP:      null.NewString(n.IP.String(), n.IP.IsValid()),
				FQDN:    null.NewString(n.FQDN, len(n.FQDN) != 0),
				VLAN:    null.NewString(n.VLAN, len(n.VLAN) != 0),
				MTU:     null.NewInt(int64(n.MTU), n.MTU != 0),
				MAC:     null.NewString(n.MAC.String(), n.MAC != nil),
				Peers:   peers,
			})
			if err != nil {
				return err
			}
			nicIDs = append(nicIDs, bi.ID)
		}

		// Delete any nics that were removed
		err = s.q.NicUpsertDelete(ctx, tx, db.NicUpsertDeleteParams{
			NodeID: node.ID,
			Ids:    nicIDs,
		})
		if err != nil {
			return err
		}

		h.ID = node.ID
	}
	return tx.Commit()
}

// DeleteHosts deletes all hosts in the given nodeset.NodeSet from the data store.
func (s *SqlStore) DeleteHosts(ns *nodeset.NodeSet) error {
	return s.q.NodeDelete(context.Background(), s.rw, ns.Iterator().StringSlice())
}

func (s *SqlStore) findNodeFromParams(params db.NodeFindParams) (*model.Host, error) {
	nodeView, err := s.q.NodeFind(context.Background(), s.ro, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &nodeView.Host, nil
}

// LoadHostFromName returns the Host with the given name
func (s *SqlStore) LoadHostFromName(name string) (*model.Host, error) {
	nodeView, err := s.q.NodeFetchByName(context.Background(), s.ro, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &nodeView.Host, nil
}

// LoadHostFromID returns the Host with the given ID
func (s *SqlStore) LoadHostFromID(uid string) (*model.Host, error) {
	nodeView, err := s.q.NodeFetchByUID(context.Background(), s.ro, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &nodeView.Host, nil
}

// ResolveIPv4 returns the list of IPv4 addresses with the given FQDN
func (s *SqlStore) ResolveIPv4(fqdn string) ([]net.IP, error) {
	if len(fqdn) == 0 {
		return nil, errors.New("invalid fqdn")
	}
	fqdnString := strings.TrimSuffix(util.Normalize(fqdn), ".")
	ips := make([]net.IP, 0)

	rows, err := s.q.NodeResolve(context.Background(), s.ro, db.NodeResolveParams{FilterFQDN: 1, FQDN: fqdnString})
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		ip, _ := netip.ParsePrefix(row.IP.String)
		if ip.IsValid() {
			ips = append(ips, net.IP(ip.Addr().AsSlice()))
		}
	}

	return ips, nil
}

// ReverseResolve returns the list of FQDNs for the given IP
func (s *SqlStore) ReverseResolve(ip string) ([]string, error) {
	if len(ip) == 0 {
		return nil, errors.New("invalid ip")
	}
	fqdn := make([]string, 0)

	rows, err := s.q.NodeResolve(context.Background(), s.ro, db.NodeResolveParams{FilterIP: 1, IP: ip})
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		names := strings.Split(row.FQDN.String, ",")
		for _, name := range names {
			fqdn = append(fqdn, name)
			break
		}
	}

	return fqdn, nil
}

// LoadHostFromMAC returns the Host that has a network interface with the give MAC address
func (s *SqlStore) LoadHostFromMAC(mac string) (*model.Host, error) {
	if len(mac) == 0 {
		return nil, errors.New("invalid mac")
	}
	return s.findNodeFromParams(db.NodeFindParams{FilterMAC: 1, MAC: null.StringFrom(mac)})
}

// Hosts returns a list of all the hosts
func (s *SqlStore) Hosts() (model.HostList, error) {
	hostList := make(model.HostList, 0)
	nodes, err := s.q.NodeAll(context.Background(), s.ro)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		hostList = append(hostList, &n.Host)
	}

	return hostList, nil
}

// FindHosts returns a list of all the hosts in the given NodeSet
func (s *SqlStore) FindHosts(ns *nodeset.NodeSet) (model.HostList, error) {
	hostList := make(model.HostList, 0)
	nodes, err := s.q.NodeFindNodeset(context.Background(), s.ro, ns.Iterator().StringSlice())
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		hostList = append(hostList, &n.Host)
	}

	return hostList, nil
}

// FindTags returns a nodeset.NodeSet of all the hosts with the given tags
func (s *SqlStore) FindTags(tags []string) (*nodeset.NodeSet, error) {
	nodes := make([]string, 0)
	names, err := s.q.NodeFindTags(context.Background(), s.ro, tags)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	for _, n := range names {
		nodes = append(nodes, n.Name)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found with tags %#v:  %w", tags, store.ErrNotFound)
	}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

// MatchTags returns a nodeset.NodeSet of all the hosts with the all given tags
func (s *SqlStore) MatchTags(tags []string) (*nodeset.NodeSet, error) {
	nodes := make([]string, 0)
	names, err := s.q.NodeFindTags(context.Background(), s.ro, tags)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	match := int64(len(tags))
	for _, n := range names {
		if n.Cnt == match {
			nodes = append(nodes, n.Name)
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found with tags %#v:  %w", tags, store.ErrNotFound)
	}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

// ProvisionHosts sets all hosts in the given NodeSet to provision (true) or unprovision (false)
func (s *SqlStore) ProvisionHosts(ns *nodeset.NodeSet, provision bool) error {
	nodeID, err := s.q.NodeID(context.Background(), s.ro, ns.Iterator().StringSlice())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no nodes found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
		}
		return err
	}

	return s.q.NodeProvision(context.Background(), s.rw, db.NodeProvisionParams{
		Nodes:     nodeID,
		Provision: provision,
	})
}

// TagHosts adds tags to all hosts in the given NodeSet
func (s *SqlStore) TagHosts(ns *nodeset.NodeSet, tags []string) error {
	nodeID, err := s.q.NodeID(context.Background(), s.ro, ns.Iterator().StringSlice())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no nodes found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
		}
		return err
	}

	if len(nodeID) == 0 {
		return fmt.Errorf("no nodes found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
	}

	ctx := context.Background()
	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, t := range tags {
		tg, err := s.q.TagUpsert(ctx, tx, t)
		if err != nil {
			return err
		}

		for _, nid := range nodeID {
			err = s.q.NodeTagUpsert(ctx, tx, db.NodeTagUpsertParams{
				TagID:  tg.ID,
				NodeID: nid,
				Value:  "", // TODO support key value pairs
			})
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// UntagHosts removes tags from all hosts in the given NodeSet
func (s *SqlStore) UntagHosts(ns *nodeset.NodeSet, tags []string) error {
	nodeID, err := s.q.NodeID(context.Background(), s.ro, ns.Iterator().StringSlice())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no nodes found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
		}
		return err
	}

	tagID, err := s.q.TagID(context.Background(), s.ro, tags)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no tags found %v: %s", tags, store.ErrNotFound)
		}
		return err
	}

	return s.q.NodeTagDelete(context.Background(), s.rw, db.NodeTagDeleteParams{
		Nodes: nodeID,
		Tags:  tagID,
	})
}

// SetBootImage sets all hosts to use the BootImage with the given name
func (s *SqlStore) SetBootImage(ns *nodeset.NodeSet, name string) error {
	ctx := context.Background()

	kernelID := null.NewInt(0, false)
	if name != "" {
		kernel, err := s.q.KernelFetch(ctx, s.ro, name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("no boot kernel found %s:  %w", name, store.ErrNotFound)
			}
			return err
		}
		kernelID.SetValid(kernel.ID)
	}

	nodeID, err := s.q.NodeID(context.Background(), s.ro, ns.Iterator().StringSlice())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no nodes found with nodeset %s:  %w", ns.String(), store.ErrNotFound)
		}
		return err
	}

	return s.q.NodeBootKernel(context.Background(), s.rw, db.NodeBootKernelParams{
		Nodes:    nodeID,
		KernelID: kernelID,
	})
}

// StoreBootImage stores a boot image in the data store. If the boot image exists it is overwritten
func (s *SqlStore) StoreBootImage(image *model.BootImage) error {
	return s.StoreBootImages(model.BootImageList{image})
}

// StoreBootImages stores a list of boot images in the data store. If the boot image exists it is overwritten
func (s *SqlStore) StoreBootImages(images model.BootImageList) error {
	ctx := context.Background()
	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for idx, image := range images {
		if image.Name == "" {
			return fmt.Errorf("name required for kernel %d: %w", idx, store.ErrInvalidData)
		}

		if image.UID.IsNil() {
			image.UID, err = ksuid.NewRandom()
			if err != nil {
				return err
			}
		}
		// Upsert kernel
		kernel, err := s.q.KernelUpsert(ctx, tx, db.KernelUpsertParams{
			ID:          null.NewInt(image.ID, image.ID != 0),
			UID:         image.UID,
			Name:        image.Name,
			Path:        image.KernelPath,
			CommandLine: null.NewString(image.CommandLine, len(image.CommandLine) != 0),
			Verify:      image.Verify,
		})
		if err != nil {
			return err
		}

		// Upsert initrd
		initrdIDs := make([]int64, 0)
		for _, rd := range image.InitrdPaths {
			ird, err := s.q.InitrdUpsert(ctx, tx, db.InitrdUpsertParams{
				KernelID: kernel.ID,
				Path:     rd,
			})
			if err != nil {
				return err
			}

			initrdIDs = append(initrdIDs, ird.ID)
		}

		if len(initrdIDs) == 0 {
			initrdIDs = append(initrdIDs, 0)
		}
		// Delete any initrds that were removed
		err = s.q.InitrdUpsertDelete(ctx, tx, db.InitrdUpsertDeleteParams{
			KernelID: kernel.ID,
			Ids:      initrdIDs,
		})
		if err != nil {
			return err
		}

		templateIDs := make([]int64, 0)
		// Upsert templates
		for ttype, tmpl := range image.ProvisionTemplates {
			id, err := s.storeTemplate(tx, kernel.ID, ttype, tmpl)
			if err != nil {
				return err
			}
			templateIDs = append(templateIDs, id)
		}

		if len(templateIDs) == 0 {
			templateIDs = append(templateIDs, 0)
		}
		// Delete any templates that were removed
		err = s.q.KernelTemplateUpsertDelete(ctx, tx, db.KernelTemplateUpsertDeleteParams{
			KernelID: kernel.ID,
			Ids:      templateIDs,
		})
		if err != nil {
			return err
		}

		image.ID = kernel.ID
	}
	return tx.Commit()
}

func (s *SqlStore) storeTemplate(tx *sql.Tx, kid int64, ttype, name string) (int64, error) {
	ctx := context.Background()
	tt, err := s.q.TemplateTypeUpsert(ctx, tx, db.TemplateTypeUpsertParams{
		Name:    ttype,
		UriName: ttype,
	})
	if err != nil {
		return 0, err
	}

	t, err := s.q.TemplateUpsert(ctx, tx, db.TemplateUpsertParams{
		Name:           name,
		TemplateTypeID: tt.ID,
	})
	if err != nil {
		return 0, err
	}

	err = s.q.KernelTemplateUpsert(ctx, tx, db.KernelTemplateUpsertParams{
		KernelID:   kid,
		TemplateID: t.ID,
	})
	if err != nil {
		return 0, err
	}

	return t.ID, nil
}

// DeleteBootImages deletes boot images from the data store.
func (s *SqlStore) DeleteBootImages(names []string) error {
	return s.q.KernelDelete(context.Background(), s.rw, names)
}

// LoadBootImage returns a BootImage with the given name
func (s *SqlStore) LoadBootImage(name string) (*model.BootImage, error) {
	kernel, err := s.q.KernelFetch(context.Background(), s.ro, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &kernel.Image, nil
}

// BootImages returns a list of all boot images
func (s *SqlStore) BootImages() (model.BootImageList, error) {
	imageList := make(model.BootImageList, 0)
	kernels, err := s.q.KernelAll(context.Background(), s.ro)
	if err != nil {
		return nil, err
	}
	for _, k := range kernels {
		imageList = append(imageList, &k.Image)
	}

	return imageList, nil
}

// RestoreFrom restores the database using the provided data dump
func (s *SqlStore) RestoreFrom(data model.DataDump) error {
	ctx := context.Background()
	for _, user := range data.Users {
		_, err := s.q.UserCreate(ctx, s.rw, db.UserCreateParams{
			Username:     user.Username,
			Role:         user.Role,
			PasswordHash: string(user.PasswordHash),
		})
		if err != nil {
			return err
		}
	}

	err := s.StoreBootImages(data.Images)
	if err != nil {
		return err
	}

	return s.StoreHosts(data.Hosts)
}

func (s *SqlStore) GetRolesByRoute(method, path string) (*[]string, error) {
	ctx := context.Background()
	roles, err := s.q.RoleFetchByMethodAndPath(ctx, s.ro, db.RoleFetchByMethodAndPathParams{
		Method: method,
		Path:   path,
	})
	if err != nil {
		return nil, err
	}

	return &roles, nil
}

func (s *SqlStore) GetRoles() (model.RoleViewList, error) {
	ctx := context.Background()
	res, err := s.q.RoleFetchView(ctx, s.ro)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *SqlStore) GetRolesByName(name string) (*model.RoleView, error) {
	ctx := context.Background()
	res, err := s.q.RoleFetchViewByName(ctx, s.ro, name)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *SqlStore) GetPermissions() (model.PermissionList, error) {
	ctx := context.Background()
	res, err := s.q.RoleFetchPermissions(ctx, s.ro)
	if err != nil {
		return nil, err
	}
	var output model.PermissionList
	for _, p := range res {
		output = append(output, model.Permission{
			Method: p.Method,
			Path:   p.Path,
		})
	}

	return output, nil
}

func (s *SqlStore) AddRole(role, inheritedRole string) error {
	ctx := context.Background()

	roleId, err := s.q.RoleAdd(ctx, s.rw, role)
	if err != nil {
		return err
	}

	if inheritedRole == "" {
		return err
	}

	inheritedRoleId, err := s.q.RoleFetchId(ctx, s.ro, inheritedRole)
	if err != nil {
		return err
	}
	permissionIds, err := s.q.RoleFetchPermissionsByRole(ctx, s.ro, inheritedRoleId)

	for _, id := range permissionIds {
		err := s.q.RoleUpsertPermission(ctx, s.rw, db.RoleUpsertPermissionParams{
			RoleID:       roleId,
			PermissionID: id,
		})
		if err != nil {
			return err
		}
	}

	return err
}

func (s *SqlStore) DeleteRole(roles []string) error {
	ctx := context.Background()

	for _, r := range roles {
		err := s.q.RoleDelete(ctx, s.rw, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SqlStore) UpdateRolePermissions(role string, permissions model.PermissionList) error {
	ctx := context.Background()
	roleId, err := s.q.RoleFetchId(ctx, s.ro, role)
	if err != nil {
		return err
	}

	var upsertedIds []int64
	for _, p := range permissions {
		permissionId, err := s.q.RoleFetchPermissionId(ctx, s.ro, db.RoleFetchPermissionIdParams{
			Method: p.Method,
			Path:   p.Path,
		})
		if err != nil {
			return err
		}

		err = s.q.RoleUpsertPermission(ctx, s.rw, db.RoleUpsertPermissionParams{
			RoleID:       roleId,
			PermissionID: permissionId,
		})
		if err != nil {
			return err
		}
		upsertedIds = append(upsertedIds, permissionId)
	}
	if len(upsertedIds) == 0 {
		upsertedIds = append(upsertedIds, 0)
	}

	err = s.q.RoleUpsertDelete(ctx, s.rw, db.RoleUpsertDeleteParams{
		RoleID: roleId,
		Ids:    upsertedIds,
	})

	return err
}

// Close closes the SqlStore database
func (s *SqlStore) Close() error {
	s.ro.Close()
	return s.rw.Close()
}
