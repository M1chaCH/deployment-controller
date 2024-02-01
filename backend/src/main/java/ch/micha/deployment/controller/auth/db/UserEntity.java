package ch.micha.deployment.controller.auth.db;

import ch.micha.deployment.controller.auth.dto.UserReadDto;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@Getter
@Setter
@AllArgsConstructor
@NoArgsConstructor
public class UserEntity {
    private UUID id;
    private String mail;
    private String password;
    private String salt;
    private boolean admin;
    private boolean active;
    private LocalDateTime createdAt;
    private LocalDateTime lastLoginAt;

    private final List<UserPageEntity> pages = new ArrayList<>();

    public void addPages(List<UserPageEntity> toAdd) {
        pages.addAll(toAdd);
    }

    public UserReadDto asDto() {
        return new UserReadDto(
            getId(),
            getMail(),
            isAdmin(),
            isActive(),
            getPages()
        );
    }
}
