select s.is_banned,
       p.value
from users.signature s
         left join secrets.passwords p on s.iaid = p.iaid
where s.iaid = $1;