package ch.micha.deployment.controller.auth.entity.adduser;

import io.helidon.dbclient.DbColumn;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.DbRow;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class AddUserMapper implements DbMapper<AddUser> {

    @Override
    public AddUser read(DbRow row) {
        DbColumn mail = row.column("mail");
        DbColumn password = row.column("password");
        DbColumn admin = row.column("admin");

        return new AddUser(
            mail.as(String.class),
            password.as(String.class),
            admin.as(Boolean.class)
        );
    }

    @Override
    public Map<String, Object> toNamedParameters(AddUser value) {
        Map<String, Object> map = new HashMap<>(3);
        map.put("mail", value.mail());
        map.put("password", value.password());
        map.put("admin", value.admin());
        return map;
    }

    @Override
    public List<?> toIndexedParameters(AddUser value) {
        return List.of(
            value.mail(),
            value.password(),
            value.admin()
        );
    }
}
