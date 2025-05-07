# Watch versions API

Control-plane has opportunity to share information about versions changes. To use the opportunity you need to take a subscription. How to subscribe on changes describes in [Subscription](#Subscription).
After connection via websocket client gets current state of versions from control-plane. And then every time when at least one of versions changes subscribers get list of version changes from control-plane.

## Subscription

To subscribe on version changes you need just open websocket connection on url `ws://control-plane:8080/api/v2/control-plane/versions/watch`.
After successful connection your application will get message with list of versions from control-plane:
```json
{
    "state": [
        {
            "createdWhen": "2020-10-12T09:01:19.166232Z",
            "stage": "ACTIVE",
            "updatedWhen": "2020-10-12T09:01:19.166232Z",
            "version": "v1"
        }
    ]
}
``` 
When versions change your application wil get list of changes. For example:
```json
{
    "changes": [
        {
            "new": {
                "createdWhen": "2020-10-12T15:01:34.811634Z",
                "stage": "LEGACY",
                "updatedWhen": "2020-10-12T15:01:34.811634Z",
                "version": "v1"
            },
            "old": {
                "createdWhen": "2020-10-12T15:01:34.811634Z",
                "stage": "ACTIVE",
                "updatedWhen": "2020-10-12T15:01:34.811634Z",
                "version": "v1"
            }
        },
        {
            "new": {
                "createdWhen": "2020-10-12T15:03:42.109465Z",
                "stage": "ACTIVE",
                "updatedWhen": "2020-10-12T15:03:42.109465Z",
                "version": "v2"
            },
            "old": {
                "createdWhen": "2020-10-12T15:03:42.109465Z",
                "stage": "CANDIDATE",
                "updatedWhen": "2020-10-12T15:03:42.109465Z",
                "version": "v2"
            }
        }
    ]
}
```

## Java + Spring Boot Example
```java
package org.qubership.example.websocketclient;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.annotations.SerializedName;
import com.google.gson.reflect.TypeToken;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.http.HttpHeaders;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.client.WebSocketClient;
import org.springframework.web.socket.client.WebSocketConnectionManager;
import org.springframework.web.socket.client.standard.StandardWebSocketClient;
import org.springframework.web.socket.handler.TextWebSocketHandler;

import java.util.Arrays;
import java.util.List;

@SpringBootApplication
public class WebsocketClientApplication implements CommandLineRunner {

    private String token = "...";

    public static void main(String[] args) {
        SpringApplication.run(WebsocketClientApplication.class, args);
    }

    @Override
    public void run(String... args) throws Exception {
        WebSocketClient client = new StandardWebSocketClient();
        CustomTextMessageHandler textMessageHandler = new CustomTextMessageHandler();
        WebSocketConnectionManager manager = new WebSocketConnectionManager(client, textMessageHandler, "ws://localhost:8080/api/v2/control-plane/versions/watch");
        HttpHeaders headers = new HttpHeaders();
        headers.add("Authorization", "Bearer " + token);
        manager.setHeaders(headers);
        manager.start();
    }

    static class CustomTextMessageHandler extends TextWebSocketHandler {

        @Override
        public void afterConnectionEstablished(WebSocketSession session) throws Exception {
            super.afterConnectionEstablished(session);
        }

        @Override
        protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
            super.handleTextMessage(session, message);
            final String pl = message.getPayload();
            GsonBuilder builder = new GsonBuilder();
            Gson gson = builder.create();
            Message msg = gson.fromJson(pl, Message.class);
            System.out.println(msg);
        }
    }

    static class Version {
        private String version;
        private String stage;

        @Override
        public String toString() {
            return "Version{" +
                    "version='" + version + '\'' +
                    ", stage='" + stage + '\'' +
                    '}';
        }
    }

    static class Change {
        @SerializedName("new")
        private Version newVersion;
        private Version old;

        @Override
        public String toString() {
            return "Change{" +
                    "new=" + newVersion +
                    ", old=" + old +
                    '}';
        }
    }

    static class Message {
        private List<Change> changes;
        private List<Version> state;

        @Override
        public String toString() {
            return "Message{" +
                    "changes=" + changes +
                    ", state=" + state +
                    '}';
        }
    }
}
```