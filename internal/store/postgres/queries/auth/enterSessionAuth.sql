with info as (select i.iaid,
                     i.ip,
                     i.device_name,
                     i.device_platform,
                     us.eid
              from jsonb_to_recordset($1) i
                       (iaid text,
                        ip text,
                        device_name text,
                        device_platform text)
                       left join users.signature us on
                  i.iaid = us.iaid)
insert
into users.entry_sessions (iaid, eid, dt, ip, device_name, device_platform)
select iaid,
       eid,
       now()::timestamptz,
       ip,
       device_name,
       device_platform
from info