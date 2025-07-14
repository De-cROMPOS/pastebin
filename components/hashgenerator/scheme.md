```mermaid
sequenceDiagram
  Client->>HashGenerator: gRPC GetHash(text="hello")
  HashGenerator->>DB Pool: Запрос соединения
  DB Pool-->>HashGenerator: Возвращает connection
  HashGenerator->>DB: INSERT INTO hash_table VALUES (...)
  HashGenerator->>DB Pool: Возвращает соединение
  HashGenerator-->>Client: Возвращает хэш
```