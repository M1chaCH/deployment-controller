package ch.micha.deployment.controller.auth.entity.page;

import io.helidon.dbclient.DbColumn;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.DbRow;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PageMapper implements DbMapper<Page> {

    @Override
    public Page read(DbRow row) {
        DbColumn id = row.column("id");
        DbColumn url = row.column("url");
        DbColumn title = row.column("title");
        DbColumn description = row.column("description");
        DbColumn privateAccess = row.column("private_access");

        return new Page(
            id.as(Integer.class),
            url.as(String.class),
            title.as(String.class),
            description.as(String.class),
            privateAccess.as(Boolean.class)
        );
    }

    @Override
    public Map<String, Object> toNamedParameters(Page value) {
        Map<String, Object> map = new HashMap<>(5);
        map.put("id", value.id());
        map.put("url", value.url());
        map.put("title", value.title());
        map.put("description", value.description());
        map.put("private_access", value.privateAccess());
        return map;
    }

    @Override
    public List<?> toIndexedParameters(Page value) {
        return List.of(
            value.id(),
            value.url(),
            value.title(),
            value.description(),
            value.privateAccess()
        );
    }
}
