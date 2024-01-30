package ch.micha.deployment.controller.auth.resource;

import ch.micha.deployment.controller.auth.db.CachedUserDb;
import ch.micha.deployment.controller.auth.db.CachedPageDb;
import ch.micha.deployment.controller.auth.db.PageEntity;
import ch.micha.deployment.controller.auth.error.BadRequestException;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import io.helidon.common.http.Http.Status;
import io.helidon.webserver.Handler;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.ServerRequest;
import io.helidon.webserver.ServerResponse;
import io.helidon.webserver.Service;
import java.sql.SQLException;
import java.util.List;
import java.util.logging.Level;
import java.util.logging.Logger;

public class PageResource implements Service{
    private static final Logger LOGGER = Logger.getLogger(PageResource.class.getSimpleName());

    private final CachedPageDb db;
    private final CachedUserDb userDb;

    public PageResource(CachedPageDb db, CachedUserDb userDb) {
        this.db = db;
        this.userDb = userDb;
    }

    @Override
    public void update(Rules rules) {
        rules
            .get("/", this::getPages)
            .post("/", Handler.create(PageEntity.class, this::createPage))
            .put("/", Handler.create(PageEntity.class, this::editPage))
            .delete("/{id}", this::deletePage);
    }

    private void getPages(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} loading all pages", new Object[]{ requestId });

        try {
            List<PageEntity> pages = db.selectPages();

            response.send(pages);
            LOGGER.log(Level.FINE, "{0} - successfully loaded all pages", requestId);
        } catch (SQLException e) {
            throw new BadRequestException("unexpected db error", "could not load pages", e);
        }
    }

    private void createPage(ServerRequest request, ServerResponse response, PageEntity toAdd) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} adding page at {1}", new Object[]{ requestId, toAdd.getUrl() });

        try {
            db.insertPage(toAdd.getId(), toAdd.getUrl(), toAdd.getTitle(), toAdd.getDescription(), toAdd.isPrivateAccess());
            userDb.invalidateCache();

            LOGGER.log(Level.FINE, "{0} added page", new Object[]{ requestId });
            response.status(Status.NO_CONTENT_204);
            response.send();
        } catch (SQLException e) {
            throw new BadRequestException("unexpected db exception", "could not insert page, db error: " + e.getMessage(), e);
        }
    }

    private void editPage(ServerRequest request, ServerResponse response, PageEntity toEdit) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        LOGGER.log(Level.FINE, "{0} updating page {1}-{2}", new Object[]{ requestId, toEdit.getId(), toEdit.getUrl() });

        try {
            db.updatePage(toEdit.getId(), toEdit.getUrl(), toEdit.getTitle(), toEdit.getDescription(), toEdit.isPrivateAccess());
            userDb.invalidateCache();

            LOGGER.log(Level.FINE, "{0} edited page", new Object[]{ requestId });
            response.status(Status.NO_CONTENT_204);
            response.send();
        } catch (SQLException e) {
            throw new BadRequestException("unexpected db exception", "could not edit page, db error: " + e.getMessage(), e);
        }
    }

    private void deletePage(ServerRequest request, ServerResponse response) {
        final String requestId = RequestLogHandler.parseRequestId(request);
        String pageId = request.path().param("id");

        LOGGER.log(Level.FINE, "{0} deleting page with id {1}", new Object[]{ requestId, pageId });

        try {
            db.deletePage(pageId);
            response.status(Status.NO_CONTENT_204);
            response.send();
        } catch (SQLException e) {
            throw new BadRequestException("unexpected db error", "could not delete page, db error: " + e.getMessage(), e);
        }
    }
}
