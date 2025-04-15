package api

import (
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	ignoredPermissions = model.PermissionList{
		model.Permission{Method: "POST", Path: "/v1/auth/signup"},
		model.Permission{Method: "POST", Path: "/v1/auth/signin"},
		model.Permission{Method: "DELETE", Path: "/v1/auth/signout"},
	}
)

// TestAccessControl loops through the generated OpenAPI spec and tries to find
// any methods + paths not inserted into the DB. If this test fails you must add
// the missing permissions via a migration file or add it to the ignored list if
// it should not be restricted by a permission.
//
// Remember to also insert into role_permission for the applicable roles so
// existing installs gain the permission.
func TestAccessControlRoutes(t *testing.T) {
	db, err := sqlstore.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("../../api/openapi.json")
	if err != nil {
		t.Fatal(err)
	}

	var perms model.PermissionList
	for _, p := range doc.Paths.Map() {
		for _, o := range p.Operations() {
			str := strings.Split(o.OperationID, "_")
			if len(str) != 2 {
				t.Fatalf("invalid fuego operation id: %s. Expected: METHOD_PATH", o.OperationID)
			}
			perms = append(perms, model.Permission{
				Method: str[0],
				Path:   str[1],
			})
		}
	}
	dbPermissionList, err := db.GetPermissions()
	if err != nil {
		t.Fatal(err)
	}

	var notInDb model.PermissionList
	for _, oaip := range perms {
		found := false

		for _, dbp := range dbPermissionList {
			// replace path params
			re, err := regexp.Compile(":[a-z]*")
			if err != nil {
				t.Fatal(err)
			}
			sqlPath := re.ReplaceAllString(oaip.Path, "%")
			if dbp.Method == oaip.Method && dbp.Path == sqlPath {
				found = true
			}

			// Check ignored permissions
			for _, ip := range ignoredPermissions {
				if oaip.Method == ip.Method && oaip.Path == ip.Path {
					found = true
				}

			}
		}

		if !found {
			notInDb = append(notInDb, oaip)
		}
	}
	if len(notInDb) > 0 {
		t.Errorf("Error: Missing permissions found in the OpenAPI spec: %#v", notInDb)
	}
}
