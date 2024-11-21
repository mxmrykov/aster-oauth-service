with external_info as (select iaid,
                              eaid,
                              name,
                              login,
                              phone,
                              password
                       from jsonb_to_recordset($1) _
                                (iaid text,
                                 eaid bigint,
                                 name text,
                                 login text,
                                 phone text,
                                 password text)),
     internal_info as (select ip,
                              device_name,
                              device_platform
                       from jsonb_to_recordset($2) _
                                (ip text,
                                 device_name text,
                                 device_platform text)),
     signature as (
         insert into
             users.signature (iaid, eid, name, login, phone, is_banned, signup_dt)
             values (external_info.iaid,
                     external_info.eaid,
                     external_info.name,
                     external_info.login,
                     external_info.phone,
                     false,
                     now()::timestamptz)),
     details as (
         insert into
             users.details (eid, profile_pics, last_online_dt, is_online_hidden)
             values (external_info.eaid,
                     '',
                     now()::timestamptz,
                     false)),
     entry_sessions as (
         insert into
             users.entry_sessions (iaid, eid, dt, ip, device_name, device_platform)
             values (external_info.iaid,
                     external_info.eaid,
                     now()::timestamptz,
                     internal_info.ip,
                     internal_info.device_name,
                     internal_info.device_platform))
insert
into secrets.passwords (iaid, value, last_reset_dt)
values (external_info.iaid, external_info.password, now()::timestamptz)