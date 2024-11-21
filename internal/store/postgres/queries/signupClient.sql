with cl as (select iaid,
                   client_id,
                   client_secret
            from jsonb_to_recordset($1) _
                     (iaid text,
                      client_id text,
                      client_secret text))
insert
into profiles.secrets (iaid, clientid, clientsecret, update_dt)
values (cl.iaid,
        cl.client_id,
        cl.client_secret,
        now()::timestamptz);