-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

drop index event_node_open;

alter table event add column counts text not null default '{}';

-- the schema this goes back to lists a finished event with the counts stored on it, so they are counted from its nodes once here
update event
set counts = (
  select json_object(
    'total', count(*),
    'pending', coalesce(sum(n.state = 'pending'), 0),
    'running', coalesce(sum(n.state = 'running'), 0),
    'succeeded', coalesce(sum(n.state = 'succeeded'), 0),
    'failed', coalesce(sum(n.state = 'failed'), 0),
    'stalled', coalesce(sum(n.state = 'stalled'), 0)
  )
  from event_node as n
  where n.event_id = event.id
)
where exists (select 1 from event_node as n where n.event_id = event.id);
