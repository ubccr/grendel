-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

alter table event drop column counts;

-- every report of a node finishing asks whether its event still has a node going, which this answers without reading the event's other rows
create index event_node_open on event_node(event_id) where state in ('pending', 'running');
