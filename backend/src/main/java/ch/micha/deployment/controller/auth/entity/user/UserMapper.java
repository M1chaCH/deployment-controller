package ch.micha.deployment.controller.auth.entity.user;

import io.helidon.dbclient.DbColumn;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.DbRow;
import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class UserMapper implements DbMapper<User> {

    @Override
    public User read(DbRow row) {
        DbColumn id = row.column("id");
        DbColumn mail = row.column("mail");
        DbColumn password = row.column("password");
        DbColumn salt = row.column("salt");
        DbColumn admin = row.column("admin");
        DbColumn createdAt = row.column("created_at");
        DbColumn lastLoginAt = row.column("last_login");

        return new User(
            id.as(Integer.class),
            mail.as(String.class),
            password.as(String.class),
            salt.as(String.class),
            admin.as(Boolean.class),
            createdAt.as(LocalDateTime.class),
            lastLoginAt.as(LocalDateTime.class)
        );
    }

    @Override
    public Map<String, Object> toNamedParameters(User value) {
        Map<String, Object> map = new HashMap<>(7);
        map.put("id", value.id());
        map.put("mail", value.mail());
        map.put("password", value.password());
        map.put("salt", value.salt());
        map.put("admin", value.admin());
        map.put("created_at", value.createdAt());
        map.put("last_login_at", value.lastLoginAt());
        return map;
    }

    @Override
    public List<?> toIndexedParameters(User value) {
        return List.of(
            value.id(),
            value.mail(),
            value.password(),
            value.salt(),
            value.admin(),
            value.createdAt(),
            value.lastLoginAt()
        );
    }
}
