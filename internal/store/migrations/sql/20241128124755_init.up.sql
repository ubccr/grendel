-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

create table node_type (
  id         integer primary key,
  name       text    not null
);

create table template_type (
  id         integer primary key,
  name       text    not null unique,
  uri_name   text    not null unique
);

create table user (
  id            integer primary key,
  username      text    not null unique,
  role          text    not null,
  password_hash text    not null,
  enabled       integer default false not null,
  created_at    timestamp default current_timestamp not null,
  updated_at    timestamp default current_timestamp not null
);

create table arch (
  id         integer primary key,
  name       text    not null
);

insert into arch (name) values ('x86_64');
insert into arch (name) values ('aarch64');

create table tag (
  id         integer primary key,
  key        text    not null unique
);

create table kernel (
  id            integer primary key,
  uid           text not null unique,
  name          text not null unique,
  version       text not null,
  path          text not null,
  arch_id       integer,
  command_line  text,
  verify        integer default false not null,
  created_at    timestamp default current_timestamp not null,
  updated_at    timestamp default current_timestamp not null
);

create table initrd (
  id            integer primary key,
  kernel_id     integer not null,
  path          text not null,
  created_at    timestamp default current_timestamp not null,
  updated_at    timestamp default current_timestamp not null,
  foreign key (kernel_id) references kernel(id) on delete cascade,
  unique(kernel_id, path)
);

create index initrd_kernel_id_idx on initrd(kernel_id);

create table template (
  id                integer primary key,
  template_type_id  integer not null,
  name              text not null unique,
  created_at        timestamp default current_timestamp not null,
  updated_at        timestamp default current_timestamp not null,
  foreign key (template_type_id) references template_type(id)
);

create table kernel_template (
  kernel_id     integer not null,
  template_id  integer not null,
  foreign key (kernel_id)     references kernel(id) on delete cascade,
  foreign key (template_id)  references template(id) on delete cascade,
  primary key (kernel_id, template_id)
);

create table node (
  id            integer primary key,
  uid           text    not null unique,
  name          text    not null unique,
  provision     integer default false not null,
  arch_id       integer,
  kernel_id     integer,
  node_type_id  integer,
  firmware      text,
  created_at    timestamp default current_timestamp not null,
  updated_at    timestamp default current_timestamp not null,
  foreign key (node_type_id) references node_type(id),
  foreign key (arch_id)      references arch(id),
  foreign key (kernel_id)    references kernel(id)
);

create index node_kernel_id_idx on node(kernel_id);
create index node_name_idx on node(name);
create index node_uid_idx on node(uid);

create table node_tag (
  id         integer primary key,
  tag_id     integer not null,
  node_id    integer not null,
  value      text    not null,
  foreign key (tag_id)  references tag(id) on delete cascade,
  foreign key (node_id) references node(id) on delete cascade,
  unique (tag_id, node_id, value)
);

create index node_tag_node_id_idx on node_tag(node_id);

create table nic (
  id            integer primary key,
  node_id       integer not null,
  nic_type      text    not null,
  name          text,
  vlan          text,
  fqdn          text,
  mac           text,
  ip            text,
  peers         text,
  mtu           integer,
  foreign key (node_id)     references node(id) on delete cascade
);

create index nic_node_id_idx on nic(node_id);

create view node_view as
select
  n.id,
  n.name,
  n.uid,
  json_object(
    '_id', n.id,
    'id', n.uid,
    'name', n.name,
    'provision', n.provision,
    'boot_image', k.name,
    'firmware', n.firmware,
    'tags',
      (select json_group_array(concat_ws(':',t.key,nt.value))
       from node_tag as nt
       join tag as t
         on nt.tag_id = t.id
       where nt.node_id = n.id
      ),
    'interfaces', (
      select json_group_array(
         json_object(
           'id', nc.id,
           'ifname', nc.name,
           'fqdn', nc.fqdn,
           'vlan', nc.vlan,
           'mac', nc.mac,
           'bmc', iif(nc.nic_type == 'bmc', true, false),
           'ip', nc.ip
         ))
       from nic as nc
       where nc.node_id = n.id and nc.nic_type != 'bond'
    ),
    'bonds', (
      select json_group_array(
         json_object(
           'id', nc.id,
           'ifname', nc.name,
           'fqdn', nc.fqdn,
           'vlan', nc.vlan,
           'mac', nc.mac,
           'peers', json_extract(nc.peers, '$'),
           'bmc', iif(nc.nic_type == 'bmc', true, false),
           'ip', nc.ip
         ))
       from nic as nc
       where nc.node_id = n.id and nc.nic_type = 'bond'
    )
  ) as host_json
from
    node as n
left join kernel as k
on n.kernel_id = k.id
;

create view kernel_view as
select
  k.id,
  k.name,
  json_object(
    '_id', k.id,
    'id', k.uid,
    'name', k.name,
    'kernel', k.path,
    'cmdline', k.command_line,
    'verify', json_extract(k.verify),
    'initrd', (
      select json_group_array(
         rd.path
       )
       from initrd as rd
       where rd.kernel_id = k.id
    ),
    'provision_templates', (
      select json_group_object(tt.uri_name, t.name)
      from kernel_template as kt
      join template t
         on kt.template_id = t.id
      join template_type tt
         on t.template_type_id = tt.id and tt.name != 'butane' and tt.name != 'user_data'
      where kt.kernel_id = k.id
    ),
    'butane', (
      select t.name
      from kernel_template as kt
      join template t
         on kt.template_id = t.id
      join template_type tt
         on t.template_type_id = tt.id and tt.name = 'butane'
      where kt.kernel_id = k.id
    ),
    'user_data', (
      select t.name
      from kernel_template as kt
      join template t
         on kt.template_id = t.id
      join template_type tt
         on t.template_type_id = tt.id and tt.name = 'user_data'
      where kt.kernel_id = k.id
    )
  ) as image_json
from
    kernel as k
;
