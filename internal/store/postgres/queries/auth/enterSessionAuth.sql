with info as (select i.iaid,
                     i.ip,
                     i.device_name,
                     i.device_platform,
                     i.signature
              from jsonb_to_recordset($1) i
                       (iaid text,
                        ip text,
                        signature text,
                        device_name text,
                        device_platform text))
insert
into users.entry_sessions (iaid, signature, dt, ip, device_name, device_platform)
select iaid,
       signature,
       now()::timestamptz,
       ip,
       device_name,
       device_platform
from info