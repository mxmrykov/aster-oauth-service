select exists(select * from profiles.secrets where clientid = $1 and clientsecret = $2 and iaid = $3) e;