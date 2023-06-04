package ch.micha.deployment.controller.auth.entity.edituser;

import io.helidon.dbclient.DbColumn;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.DbRow;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class EditUserMapper implements DbMapper<EditUser> {

    @Override
    public EditUser read(DbRow row) {
        DbColumn id = row.column("id");
        DbColumn mail = row.column("mail");
        DbColumn password = row.column("password");
        DbColumn admin = row.column("admin");
        DbColumn viewPrivate = row.column("view_private");

        return new EditUser(
            id.as(Integer.class),
            mail.as(String.class),
            password.as(String.class),
            admin.as(Boolean.class),
            viewPrivate.as(Boolean.class)
        );
    }

    @Override
    public Map<String, Object> toNamedParameters(
        EditUser value) {
        Map<String, Object> map = new HashMap<>(5);
        map.put("id", value.id());
        map.put("mail", value.mail());
        map.put("password", value.password());
        map.put("admin", value.admin());
        map.put("view_private", value.viewPrivate());
        return map;
    }

    @Override
    public List<?> toIndexedParameters(EditUser value) {
        return List.of(
            value.id(),
            value.mail(),
            value.password(),
            value.admin(),
            value.viewPrivate()
        );
    }
}
