package ch.micha.deployment.controller.auth;

import ch.micha.deployment.controller.auth.auth.AuthHandler;
import ch.micha.deployment.controller.auth.auth.AuthService;
import ch.micha.deployment.controller.auth.resource.PageResource;
import ch.micha.deployment.controller.auth.resource.UserResource;
import io.helidon.common.LogConfig;
import io.helidon.common.reactive.Single;
import io.helidon.config.Config;
import io.helidon.dbclient.DbClient;
import io.helidon.media.jsonb.JsonbSupport;
import io.helidon.media.jsonp.JsonpSupport;
import io.helidon.openapi.OpenAPISupport;
import io.helidon.webserver.Routing;
import io.helidon.webserver.WebServer;
import java.util.logging.Level;
import java.util.logging.Logger;

public final class Main {
    private static final Logger LOGGER = Logger.getLogger(Main.class.getSimpleName());

    private Main() {
    }

    public static void main(final String[] args) {
        startServer();
    }

    /**
     * Start the server.
     *
     * @return the created {@link WebServer} instance
     */
    static Single<WebServer> startServer() {
        // load logging configuration
        LogConfig.configureRuntime();

        // By default, this will pick up application.yaml from the classpath
        Config config = Config.create();

        WebServer server = WebServer.builder(createRouting(config))
            .config(config.get("server"))
            .addMediaSupport(JsonpSupport.create())
            .addMediaSupport(JsonbSupport.create())
            .build();

        Single<WebServer> webserver = server.start();

        webserver.forSingle(ws -> {
            LOGGER.log(Level.INFO, "server up and running at http://localhost:{0}", ws.port());
            ws.whenShutdown().thenRun(() -> LOGGER.info("server shut down! enjoy your resources (:"));
        }).exceptionallyAccept(t -> LOGGER.log(Level.SEVERE, "failed to start server:", t));

        return webserver;
    }

    /**
     * Creates new {@link Routing}.
     *
     * @param config configuration of this server
     * @return routing configured with JSON support, a health check, and a service
     */
    private static Routing createRouting(Config config) {
        Config dbConfig = config.get("db");
        DbClient dbClient = DbClient.builder(dbConfig).build();

        Config securityConfig = config.get("app.security");
        AuthService authService = new AuthService(dbClient, securityConfig);
        AuthHandler authHandler = new AuthHandler(authService);
        UserResource userResource = new UserResource(dbClient);
        PageResource pageResource = new PageResource(dbClient);

        Routing.Builder builder = Routing.builder()
            .register(OpenAPISupport.create(config))
            .register("/security", authService)
            .any("/users", authHandler)
            .register("/users", userResource)
            .any("/pages", authHandler)
            .register("/pages", pageResource);

        return builder.build();
    }
}
