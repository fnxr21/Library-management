# Task

**Challenge :**

As a Backend Engineer you are asked to create a Backend system for a library management application. The applications that will be created include book management, book categories and user management. You don't need to create a Frontend but just focus on creating a RESTful API.

**Goals :**

- Building a scalable backend system.
- Using microservice architecture.
- Handle race conditions and ensure data consistency.
- Implementation of data replication and failover.

**Specifications :**

1. Programing language: Golang
2. Database: PostgreSQL (each service has its own database)
3. Pattern: Microservice
4. Communication between services: gRPC
5. Middleware: JWT untuk autentikasi dan otorisasi
6. Containerization: Docker or Kubernetes

**MVP Features :**

1. Book Management
2. Author Management
3. Book Categories Management
4. Book Stock Management
5. Borrowing and Returning Books
6. Search and Recommendation feature

**Microservices with multiple database :**

- BookService and Book Database
- AuthorService and Author Database
- CategoryService and Category Database
- UserService/AuthService and User Database

**Assessment Criteria :**

1. System Design
   - Good README Documentation.
   - RESTful API Design.
   - Entity Relationship Diagram.
2. Development
   - Database Implementation.
   - Communication between microservices.
   - Error handling and Logging.
   - Unit test.
   - Using cache like Redis is a plus.
3. Deployment
   - Dockerfile and docker-compose file
   - Upload docker image to docker registry (docker hub)
   - Deploy app to cloud.

**Rules :**

1. The challenge must be submitted in **4 days** after you get this challenge.
2. Save the code you created to the private Github repository, and invite username **backend-syn (<https://github.com/backend-syn>)** to the repository and then collect the repository link in the form we have provided.
3. Also collect the project link that has been built and saved to the registry container [hub.docker.com](https://hub.docker.com/) and the api spec link to the form we have provided.
4. **If you can complete the challenge test before the deadline, it will be taken into consideration by us.**
