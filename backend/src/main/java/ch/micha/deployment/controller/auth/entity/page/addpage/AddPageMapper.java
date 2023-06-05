package ch.micha.deployment.controller.auth.entity.page.addpage;

import io.helidon.dbclient.DbColumn;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.DbRow;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class AddPageMapper implements DbMapper<AddPage> {

    @Override
    public AddPage read(DbRow row) {
        DbColumn url = row.column("url");
        DbColumn title = row.column("title");
        DbColumn description = row.column("description");
        DbColumn privateAccess = row.column("private_access");

        return new AddPage(
            url.as(String.class),
            title.as(String.class),
            description.as(String.class),
            privateAccess.as(Boolean.class)
        );
    }

    @Override
    public Map<String, Object> toNamedParameters(AddPage value) {
        Map<String, Object> map = new HashMap<>(4);
        map.put("url", value.url());
        map.put("title", value.title());
        map.put("description", value.description());
        map.put("private_access", value.privateAccess());
        return map;
    }

    @Override
    public List<?> toIndexedParameters(AddPage value) {
        return List.of(
            value.url(),
            value.title(),
            value.description(),
            value.privateAccess()
        );
    }
}
