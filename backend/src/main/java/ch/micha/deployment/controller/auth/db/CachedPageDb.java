package ch.micha.deployment.controller.auth.db;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.logging.Level;
import java.util.logging.Logger;

public class CachedPageDb {
    private static final Logger LOGGER = Logger.getLogger(CachedPageDb.class.getSimpleName());

    protected final Connection db;
    protected final Map<String, PageEntity> cache = new HashMap<>();

    public CachedPageDb(Connection db) {
        this.db = db;
    }

    public List<PageEntity> selectPages() throws SQLException {
        PreparedStatement pagesStatement = db.prepareStatement("""
            select id, url, title, description, private_page
            from pages
            order by pages.id
            """);
        ResultSet pagesResult = pagesStatement.executeQuery();

        List<PageEntity> pages = new ArrayList<>();
        while (pagesResult.next()) {
            PageEntity page = parsePage(pagesResult);
            cache.put(page.getId(), page);
            pages.add(page);
        }

        pagesResult.close();
        pagesStatement.close();
        return pages;
    }

    public Optional<PageEntity> selectPage(String pageId) {
        try {
            PageEntity page = cache.get(pageId);
            if(page != null)
                return Optional.of(page);

            PreparedStatement pageStatement = db.prepareStatement("""
                select id, url, title, description, private_page
                from pages
                where id = ?
                """);
            pageStatement.setString(1, pageId);

            ResultSet pageResult = pageStatement.executeQuery();
            if(!pageResult.next()) {
                return Optional.empty();
            }

            page = parsePage(pageResult);
            cache.put(page.getId(), page);
            pageResult.close();
            pageStatement.close();
            return Optional.of(page);
        } catch (SQLException e) {
            LOGGER.log(Level.WARNING, "could not select page by id: %s".formatted(pageId), e);
            return Optional.empty();
        }
    }

    protected PageEntity parsePage(ResultSet pageResult) throws SQLException {
        return new PageEntity(
            pageResult.getString("id"),
            pageResult.getString("url"),
            pageResult.getString("title"),
            pageResult.getString("description"),
            pageResult.getBoolean("private_page")
            );
    }

    public void insertPage(String pageId, String url, String title, String description, boolean privateAccess) throws SQLException {
        PreparedStatement insertPage = db.prepareStatement("""
            insert into pages (id, url, title, description, private_page)
            values (?, ?, ?, ?, ?)
            """);
        insertPage.setString(1, pageId);
        insertPage.setString(2, url);
        insertPage.setString(3, title);
        insertPage.setString(4, description);
        insertPage.setBoolean(5, privateAccess);

        insertPage.execute();
        insertPage.close();

        cache.put(pageId, new PageEntity(pageId, url, title, description, privateAccess));
    }

    public void updatePage(String pageId, String url, String title, String description, boolean privateAccess) throws SQLException {
        PreparedStatement updatePage = db.prepareStatement("""
            update pages set url = ?, title = ?, description = ?, private_page = ?
            where id = ?
            """);
        updatePage.setString(1, url);
        updatePage.setString(2, title);
        updatePage.setString(3, description);
        updatePage.setBoolean(4, privateAccess);
        updatePage.setString(5, pageId);

        updatePage.executeUpdate();
        updatePage.close();

        cache.put(pageId, new PageEntity(pageId, url, title, description, privateAccess));
    }

    public void deletePage(String pageId) throws SQLException {
        PreparedStatement deletePage = db.prepareStatement("""
            delete from pages where id = ?
            """);
        deletePage.setString(1, pageId);

        deletePage.executeUpdate();
        deletePage.close();

        cache.remove(pageId);
    }
}
