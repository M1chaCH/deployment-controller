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
    public static final String CLAIM_USER_MAIL = "user_mail";
    public static final String CLAIM_ADMIN = "admin";
    public static final String CLAIM_PRIVATE_ACCESS = "private_access";

    private String issuer;
    private Date issuedAt;
    private String userMail;
    private boolean admin;
    private boolean privateAccess;
    private Date expiresAt;
}
