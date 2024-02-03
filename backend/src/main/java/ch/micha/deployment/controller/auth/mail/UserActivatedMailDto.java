package ch.micha.deployment.controller.auth.mail;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class UserActivatedMailDto {
    private String userMail;
    private String time;
}
