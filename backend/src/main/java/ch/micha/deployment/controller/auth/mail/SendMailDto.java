package ch.micha.deployment.controller.auth.mail;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class SendMailDto {
    private Type mailType;
    private Object data;
    private String recipient;

    public enum Type {
        LOGIN_GRANT,
        PAGE_INVITATION,
        USER_ACTIVATED,
        CONTACT_REQUEST,
    }
}
