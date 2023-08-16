
# Dockerized WordPress Infrastructure

Welcome to the Dockerized WordPress Infrastructure project. This project offers a highly scalable setup to host multiple WordPress instances, all managed and load-balanced by a dedicated NGINX container.

## Project Overview

- **Scalable WordPress**: We run 3 WordPress containers, ensuring your sites remain fast and responsive.
- **Persistent Data**: All WordPress data is safely stored in a containerized MySQL database, ensuring no data loss.
- **Monitoring**: The entire setup is monitored using a Telegraf plugin based on PowerTop. You can rest assured about the performance and health of your sites.
- **Metrics Visualization**: We've integrated an InfluxDB container to help you visualize various metrics, enhancing your monitoring and analytical capabilities.
- **Docker Images**: All Docker images resulting from this infrastructure's build are hosted in the GitLab Docker registry.
- **Deployment**: The entire project is seamlessly deployed using `docker-compose`.

## Docker Compose Details

Our `docker-compose.yml` configures the following services:

- **NGINX** for load balancing the 3 WordPress containers.
- **3 WordPress Containers** (`wordpress1`, `wordpress2`, and `wordpress3`).
- **MySQL** for data persistence across WordPress sites.
- **PowerTop** for enhanced monitoring using the Telegraf plugin.
- **InfluxDB** for metrics visualization.

## Setup

1. **Clone this Repository** (assuming you've hosted this on a Git platform).
   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```

2. **Deploy using Docker Compose**
   ```bash
   docker-compose up -d --build
   ```

3. **Access your Sites**
   - WordPress: [http://localhost:80](http://localhost:80)
   - InfluxDB Dashboard: [http://localhost:8086](http://localhost:8086)

## Conclusion

This project aims to provide a robust, scalable, and easily deployable WordPress infrastructure. By leveraging the power of Docker and containerization, we ensure efficiency, scalability, and ease of management.
