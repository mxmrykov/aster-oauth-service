select exists(select * from users.signature where phone = $1) as e;