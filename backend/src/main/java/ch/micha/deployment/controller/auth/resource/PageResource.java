package ch.micha.deployment.controller.auth.resource;

import ch.micha.deployment.controller.auth.entity.page.Page;
import ch.micha.deployment.controller.auth.entity.page.addpage.AddPage;
import ch.micha.deployment.controller.auth.error.AppRequestException;
import io.helidon.common.http.Http.Status;
import io.helidon.dbclient.DbClient;
import io.helidon.webserver.Handler;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import io.helidon.webserver.Service;
import jakarta.json.JsonObject;
import java.util.logging.Level;
import java.util.logging.Logger;

public class PageResource implements Service{
    private static final Logger LOGGER = Logger.getLogger(PageResource.class.getSimpleName());

    private final DbClient db;

    public PageResource(DbClient db) {
        this.db = db;
    }

    @Override
    public void update(Rules rules) {
        rules
            .get("/", this::getPages)
            .post("/", Handler.create(AddPage.class, this::createPage))
            .put("/", Handler.create(Page.class, this::editPage))
            .delete("/{id}", this::deletePage);
    }

    private void getPages(ServerRequest request, ServerResponse response) {
        LOGGER.info("loading all pages");
        response.send(db.execute(exec -> exec.namedQuery("select-pages"))
                .map(item -> item.as(JsonObject.class)), JsonObject.class)
            .thenAccept(sentResponse ->
                LOGGER.log(Level.INFO, "{0} - successfully loaded all pages", sentResponse.status()));
    }

    private void createPage(ServerRequest request, ServerResponse response, AddPage toAdd) {
        LOGGER.log(Level.INFO, "adding page at {0}", new Object[]{ toAdd.url() });
        db.execute(exec -> exec
                .createNamedInsert("insert-page")
                .namedParam(toAdd)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.INFO, "added {0} page(s)", count);
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, t));
    }

    private void editPage(ServerRequest request, ServerResponse response, Page toEdit) {
        LOGGER.log(Level.INFO, "updating page {0}-{1}", new Object[]{ toEdit.id(), toEdit.url() });

        db.execute(exec -> exec
                .createNamedUpdate("update-page")
                .namedParam(toEdit)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.INFO, "changed {0} page(s)", new Object[]{ count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, t));
    }

    private void deletePage(ServerRequest request, ServerResponse response) {
        int pageId = Integer.parseInt(request.path().param("id"));

        LOGGER.log(Level.INFO, "deleting page with id {0}", new Object[]{ pageId });

        db.execute(exec -> exec
                .createNamedDelete("delete-page")
                .addParam("id", pageId)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.INFO, "deleted {0} page(s)", new Object[]{ count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, t));
    }
}
