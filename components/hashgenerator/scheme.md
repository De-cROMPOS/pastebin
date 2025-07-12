sequenceDiagram
  Client->>Server: gRPC GetHash(text="hello")
  Server->>DB Pool: Запрос соединения
  DB Pool-->>Server: Возвращает connection
  Server->>DB: INSERT INTO hashes VALUES (...)
  Server-->>Client: Возвращает хэш
  Server->>DB Pool: Возвращает соединение