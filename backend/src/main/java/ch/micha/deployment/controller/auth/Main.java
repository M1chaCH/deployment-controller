package ch.micha.deployment.controller.auth;

import ch.micha.deployment.controller.auth.auth.AuthHandler;
import ch.micha.deployment.controller.auth.auth.AuthService;
import ch.micha.deployment.controller.auth.error.AppRequestException;
import ch.micha.deployment.controller.auth.error.GlobalErrorHandler;
import ch.micha.deployment.controller.auth.logging.RequestLogHandler;
import ch.micha.deployment.controller.auth.resource.PageResource;
import ch.micha.deployment.controller.auth.resource.UserResource;
import io.helidon.common.LogConfig;
import io.helidon.common.reactive.Single;
import io.helidon.config.Config;
import io.helidon.dbclient.DbClient;
import io.helidon.media.jsonb.JsonbSupport;
import io.helidon.media.jsonp.JsonpSupport;
import io.helidon.webserver.HttpException;
import io.helidon.webserver.Routing;
import io.helidon.webserver.WebServer;
import io.helidon.webserver.cors.CorsSupport;
import io.helidon.webserver.cors.CrossOriginConfig;
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
        Config appConfig = config.get("app");

        String clientUrl = appConfig.get("security.frontend").asString().get();
        final CorsSupport corsSupport = CorsSupport.builder()
            .addCrossOrigin(CrossOriginConfig.builder()
                .pathPattern("*")
                .allowOrigins(clientUrl)
                .allowCredentials(true)
                .allowMethods("GET", "PUT", "POST", "DELETE", "OPTIONS")
                .build())
            .build();

        Config dbConfig = config.get("db");
        DbClient dbClient = DbClient.builder(dbConfig).build();

        GlobalErrorHandler errorHandler = new GlobalErrorHandler();
        RequestLogHandler requestLogHandler = new RequestLogHandler(appConfig);
        AuthService authService = new AuthService(dbClient, appConfig);
        AuthHandler authHandler = new AuthHandler(authService);

        Routing.Builder builder = Routing.builder()
            .error(AppRequestException.class, errorHandler::handleAppRequestException)
            .error(HttpException.class, errorHandler::handleHttpException)
            .error(Exception.class, errorHandler::handleException)
            .any(requestLogHandler)
            .any(corsSupport)
            .register("/security", authService)
            .any("/users", authHandler)
            .any("/pages", authHandler)
            .register("/users", new UserResource(dbClient))
            .register("/pages", new PageResource(dbClient));

        return builder.build();
    }
}
