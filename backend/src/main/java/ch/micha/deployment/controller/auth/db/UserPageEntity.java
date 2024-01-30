package ch.micha.deployment.controller.auth.db;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;

@Getter
@Setter
@AllArgsConstructor
@RequiredArgsConstructor
public class UserPageEntity {
    private String pageId;
    private String url;
    private String title;
    private String description;
    private boolean privatePage;
    private boolean hasAccess;
}
