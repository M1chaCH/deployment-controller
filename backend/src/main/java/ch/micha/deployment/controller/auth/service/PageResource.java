package ch.micha.deployment.controller.auth.service;

import ch.micha.deployment.controller.auth.entity.page.Page;
import ch.micha.deployment.controller.auth.entity.page.addpage.AddPage;
import io.helidon.common.http.Http.Status;
import io.helidon.dbclient.DbClient;
import io.helidon.webserver.Handler;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import io.helidon.webserver.Service;
import jakarta.json.JsonObject;
import java.util.concurrent.CompletionException;
import java.util.logging.Level;
import java.util.logging.Logger;
import org.postgresql.util.PSQLException;

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
            .exceptionally(throwable -> sendError(throwable, response));
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
            .exceptionally(t -> sendError(t, response));
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
            .exceptionally(t -> sendError(t, response));
    }

    @SuppressWarnings({"java:S3516", "SameReturnValue"})
    private <T> T sendError(Throwable throwable, ServerResponse response) {
        Throwable realCause = throwable;
        if (throwable instanceof CompletionException)
            realCause = throwable.getCause();

        if(realCause instanceof PSQLException psqlException &&
            psqlException.getMessage().contains("value violates unique constraint")) {

            LOGGER.log(Level.INFO, "unique constraint was violated: {0}", psqlException.getServerErrorMessage());
            response.status(Status.BAD_REQUEST_400);
            response.send("value already exists: " + psqlException.getServerErrorMessage());
            return null;
        }

        LOGGER.log(Level.WARNING, "caught error", realCause);
        response.status(Status.INTERNAL_SERVER_ERROR_500);
        response.send("Failed to process request: " + realCause.getClass().getName() +
            "(" + realCause.getMessage() + ")");
        return null;
    }
}
