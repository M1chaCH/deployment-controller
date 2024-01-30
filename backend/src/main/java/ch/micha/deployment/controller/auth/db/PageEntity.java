package ch.micha.deployment.controller.auth.db;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@Getter
@Setter
@AllArgsConstructor
@NoArgsConstructor
public class PageEntity {
    private String id;
    private String url;
    private String title;
    private String description;
    private boolean privateAccess;
}
