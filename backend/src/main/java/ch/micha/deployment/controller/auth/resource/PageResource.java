package ch.micha.deployment.controller.auth.resource;

import ch.micha.deployment.controller.auth.entity.page.Page;
import ch.micha.deployment.controller.auth.error.AppRequestException;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
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
            .post("/", Handler.create(Page.class, this::createPage))
            .put("/", Handler.create(Page.class, this::editPage))
            .delete("/{id}", this::deletePage);
    }

    private void getPages(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} loading all pages", new Object[]{ requestId });

        response.send(db.execute(exec -> exec.namedQuery("select-pages"))
                .map(item -> item.as(JsonObject.class)), JsonObject.class)
            .thenAccept(sentResponse -> LOGGER.log(Level.FINE, "{0} - successfully loaded all pages", requestId))
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }

    private void createPage(ServerRequest request, ServerResponse response, Page toAdd) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} adding page at {1}", new Object[]{ requestId, toAdd.url() });

        db.execute(exec -> exec
                .createNamedInsert("insert-page")
                .namedParam(toAdd)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.FINE, "{0} added {1} page(s)", new Object[]{ requestId, count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }

    private void editPage(ServerRequest request, ServerResponse response, Page toEdit) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} updating page {1}-{2}", new Object[]{ requestId, toEdit.id(), toEdit.url() });

        db.execute(exec -> exec
                .createNamedUpdate("update-page")
                .namedParam(toEdit)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.FINE, "{0} changed {1} page(s)", new Object[]{ requestId, count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }

    private void deletePage(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        String pageId = request.path().param("id");

        LOGGER.log(Level.FINE, "{0} deleting page with id {1}", new Object[]{ requestId, pageId });

        db.execute(exec -> exec
                .createNamedDelete("delete-page")
                .addParam("id", pageId)
                .execute())
            .thenAccept(count -> {
                LOGGER.log(Level.FINE, "{0} deleted {1} page(s)", new Object[]{ requestId, count });
                response.status(Status.NO_CONTENT_204);
                response.send();
            })
            .exceptionally(t -> AppRequestException.respondFitting(response, requestId, t));
    }
}
