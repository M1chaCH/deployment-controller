FROM maven:3.8.4-openjdk-17-slim AS build

WORKDIR /helidon

ADD pom.xml pom.xml
RUN mvn package -Dmaven.test.skip 

ADD src src
RUN mvn package -DskipTests

RUN echo "built backend!"

FROM openjdk:17-jdk-slim
WORKDIR /helidon

# Copy the binary built in the 1st stage
COPY --from=build /helidon/target/backend.jar ./
COPY --from=build /helidon/target/libs ./libs

CMD ["java", "-jar", "backend.jar"]

EXPOSE 8080
