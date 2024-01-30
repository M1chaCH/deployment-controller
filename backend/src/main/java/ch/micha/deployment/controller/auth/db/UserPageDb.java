package ch.micha.deployment.controller.auth.db;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

public class UserPageDb {

    protected final Connection db;

    public UserPageDb(Connection db) {
        this.db = db;
    }

    public List<UserPageEntity> selectPagesForUser(UUID userId) throws SQLException {
        PreparedStatement pagesStatement = db.prepareStatement("""
            select p.id, url, title, description, private_page, up.user_id
            from pages as p
            left join public.user_page up on ? = up.user_id
            """);
        pagesStatement.setObject(1, userId);
        ResultSet pagesResult = pagesStatement.executeQuery();
        List<UserPageEntity> selectedPages = new ArrayList<>();

        while (pagesResult.next()) {
            UserPageEntity entity = new UserPageEntity();
            entity.setPageId(pagesResult.getString("id"));
            entity.setUrl(pagesResult.getString("url"));
            entity.setTitle(pagesResult.getString("title"));
            entity.setDescription(pagesResult.getString("description"));
            entity.setPrivatePage(pagesResult.getBoolean("private_page"));
            entity.setHasAccess(pagesResult.getObject("user_id") != null);
            selectedPages.add(entity);
        }
        pagesResult.close();
        pagesStatement.close();

        return selectedPages;
    }

    void insertPagesForUser(UUID userId, String... pages) throws SQLException {
        if(pages.length == 0)
            return;

        StringBuilder valuesBuilder = new StringBuilder();
        for (int i = 0; i < pages.length; i++) {
            valuesBuilder.append("(?, ?)");

            if(i < pages.length - 1)
                valuesBuilder.append(",");
        }

        PreparedStatement pageInsert = db.prepareStatement("""
            insert into user_page (user_id, page_id) values
            """ +  valuesBuilder);
        int index = 1;
        for (String page : pages) {
            pageInsert.setObject(index, userId);
            index++;
            pageInsert.setString(index, page);
            index++;
        }

        pageInsert.execute();
        pageInsert.close();
    }

    void deletePagesForUser(UUID userId, String... pages) throws SQLException {
        StringBuilder whereBuilder = new StringBuilder();
        for (int i = 0; i < pages.length; i++) {
            whereBuilder.append("user_id = ? and page_id = ?");

            if(i < pages.length - 1)
                whereBuilder.append(" or ");
        }

        PreparedStatement pageDelete = db.prepareStatement("""
            delete from user_page where 
            """ +  whereBuilder);
        int index = 1;
        for (String page : pages) {
            pageDelete.setObject(index, userId);
            index++;
            pageDelete.setString(index, page);
            index++;
        }

        pageDelete.execute();
        pageDelete.close();
    }
}
