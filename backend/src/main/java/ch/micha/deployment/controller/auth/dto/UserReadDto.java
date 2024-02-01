package ch.micha.deployment.controller.auth.dto;

import ch.micha.deployment.controller.auth.db.UserPageEntity;
import java.util.List;
import java.util.UUID;

public record UserReadDto(
    UUID userId,
    String mail,
    boolean admin,
    boolean active,
    List<UserPageEntity> pages
) {
}
