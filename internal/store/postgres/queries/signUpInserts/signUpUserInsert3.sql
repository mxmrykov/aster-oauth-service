with iinfo as (select i.ip,
                      i.device_name,
                      i.device_platform,
                      e.iaid,
                      e.eaid,
                      e.name,
                      e.login,
                      e.phone,
                      e.password
               from jsonb_to_recordset($2) i
                        (ip text,
                         device_name text,
                         device_platform text),
                    jsonb_to_recordset($1) e
                        (iaid text,
                         eaid bigint,
                         name text,
                         login text,
                         phone text,
                         password text))
insert
into users.entry_sessions (iaid, eid, dt, ip, device_name, device_platform)
select iaid,
       eaid,
       now()::timestamptz,
       ip,
       device_name,
       device_platform
from iinfo