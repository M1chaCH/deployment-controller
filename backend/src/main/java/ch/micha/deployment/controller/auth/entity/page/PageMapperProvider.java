package ch.micha.deployment.controller.auth.entity.page;

import ch.micha.deployment.controller.auth.entity.page.addpage.AddPage;
import ch.micha.deployment.controller.auth.entity.page.addpage.AddPageMapper;
import io.helidon.dbclient.DbMapper;
import io.helidon.dbclient.spi.DbMapperProvider;
import jakarta.annotation.Priority;
import java.util.Optional;

@Priority(1000)
public class PageMapperProvider implements DbMapperProvider {
    private static final PageMapper PAGE_MAPPER = new PageMapper();
    private static final AddPageMapper ADD_PAGE_MAPPER = new AddPageMapper();

    @SuppressWarnings("unchecked")
    @Override
    public <T> Optional<DbMapper<T>> mapper(Class<T> type) {
        if(type.equals(Page.class))
            return Optional.of((DbMapper<T>) PAGE_MAPPER);
        if(type.equals(AddPage.class))
            return Optional.of((DbMapper<T>) ADD_PAGE_MAPPER);

        return Optional.empty();
    }
}
