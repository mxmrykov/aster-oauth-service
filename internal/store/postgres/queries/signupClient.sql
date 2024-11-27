with cl as (select iaid,
                   client_id,
                   client_secret
            from jsonb_to_recordset($1) _
                     (iaid text,
                      client_id text,
                      client_secret text))
insert
into profiles.secrets (iaid, clientid, clientsecret, update_dt)
select iaid,
       client_id,
       client_secret,
       now()::timestamptz
from cl;