package ch.micha.deployment.controller.auth.service;

import io.helidon.dbclient.DbClient;
import io.helidon.webserver.Routing.Rules;
import io.helidon.webserver.Service;

public class AuthService implements Service {

    private final DbClient db;

    public AuthService(DbClient db) {
        this.db = db;
    }

    @Override
    public void update(Rules rules) {

    }
}
