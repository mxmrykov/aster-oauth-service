with iinfo as (select i.ip,
                      i.device_name,
                      i.device_platform,
                      e.iaid,
                      e.signature
               from jsonb_to_recordset($2) i
                        (ip text,
                         device_name text,
                         device_platform text),
                    jsonb_to_recordset($1) e
                        (iaid text,
                         signature text))
insert
into users.entry_sessions (iaid, signature, dt, ip, device_name, device_platform)
select iaid,
       signature,
       now()::timestamptz,
       ip,
       device_name,
       device_platform
from iinfo