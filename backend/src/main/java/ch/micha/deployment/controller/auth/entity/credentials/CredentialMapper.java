package ch.micha.deployment.controller.auth.entity.credentials;

import io.helidon.dbclient.DbColumn;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.DbRow;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class CredentialMapper implements DbMapper<Credential> {

    @Override
    public Credential read(DbRow row) {
        DbColumn mail = row.column("mail");
        DbColumn password = row.column("password");

        return new Credential(
            mail.as(String.class),
            password.as(String.class)
        );
    }

    @Override
    public Map<String, Object> toNamedParameters(Credential value) {
        Map<String, Object> map = new HashMap<>(2);
        map.put("mail", value.mail());
        map.put("password", value.password());
        return map;
    }

    @Override
    public List<?> toIndexedParameters(Credential value) {
        return List.of(
            value.mail(),
            value.password()
        );
    }
}
