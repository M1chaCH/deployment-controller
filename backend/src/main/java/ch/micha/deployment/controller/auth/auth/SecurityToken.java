package ch.micha.deployment.controller.auth.auth;

import java.util.Date;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class SecurityToken {
    public static final String CLAIM_USER_ID = "user_id";
    public static final String CLAIM_USER_MAIL = "user_mail";
    public static final String CLAIM_ADMIN = "admin";
    public static final String CLAIM_ACTIVE = "active";
    public static final String CLAIM_PRIVATE_ACCESS = "private_access";
    public static final String CLAIM_PRIVATE_ACCESS_DELIMITER = "&&";

    private String issuer;
    private Date issuedAt;
    private String userId;
    private String userMail;
    private boolean admin;
    private boolean active;
    private String privatePagesAccess;
    private Date expiresAt;
}
