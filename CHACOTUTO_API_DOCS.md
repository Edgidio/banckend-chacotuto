# CHACOTUTO — Documentación API Backend

> Documentación completa del backend ChacOtuto.
> Servidor: **Go (Fiber v2)** | Base de Datos: **SQLite** | Auth: **JWT**

---

## Información del Servidor

| Campo | Valor |
|---|---|
| **Puerto** | `8080` |
| **URL Local** | `http://localhost:8080` |
| **URL Red (teléfono)** | `http://<TU_IP_LOCAL>:8080` |
| **WebSocket Drones** | `ws://<HOST>:8080/ws/drone` |
| **WebSocket GCS** | `ws://<HOST>:8080/ws/gcs?token=<JWT>` |
| **Base de Datos** | `chacotuto.db` (SQLite, se crea automáticamente) |
| **Admin por defecto** | `admin` / `chacotuto2026` |

---

## Tabla de Rutas

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| `POST` | `/api/auth/login` | ❌ No | Iniciar sesión, obtener JWT |
| `POST` | `/api/auth/register` | 🔐 Admin | Crear nuevo usuario |
| `GET` | `/api/auth/me` | 🔐 JWT | Ver perfil del usuario autenticado |
| `GET` | `/api/drones` | 🔐 JWT | Listar todos los drones |
| `GET` | `/api/drones/:id` | 🔐 JWT | Detalle de un dron + sus misiones |
| `GET` | `/api/drones/:id/telemetry` | 🔐 JWT | Historial de telemetría (paginado) |
| `GET` | `/api/missions` | 🔐 JWT | Listar todas las misiones |
| `GET` | `/api/missions/:id` | 🔐 JWT | Detalle de una misión |
| `POST` | `/api/missions` | 🔐 JWT | Crear y asignar misión a un dron |
| `DELETE` | `/api/missions/:id` | 🔐 JWT | Cancelar una misión activa |
| `WS` | `/ws/drone` | ❌ No | WebSocket para drones (protocolo) |
| `WS` | `/ws/gcs?token=<JWT>` | 🔐 JWT | WebSocket para el dashboard GCS |

---

## Autenticación

Todas las rutas protegidas requieren el header:

```
Authorization: Bearer <token>
```

El token se obtiene con el endpoint de login y tiene una duración de **24 horas**.

---

## Endpoints REST

---

### 🔑 `POST /api/auth/login`

Inicia sesión y obtiene un token JWT.

**Auth requerida:** No

**Request Body:**

```json
{
  "username": "admin",
  "password": "chacotuto2026"
}
```

| Campo | Tipo | Requerido | Descripción |
|---|---|---|---|
| `username` | string | ✅ | Nombre de usuario |
| `password` | string | ✅ | Contraseña |

**Response `200 OK`:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin"
  }
}
```

**Response `401 Unauthorized`:**

```json
{
  "error": "Credenciales inválidas"
}
```

---

### 🔑 `POST /api/auth/register`

Crea un nuevo usuario. **Solo administradores pueden crear usuarios.**

**Auth requerida:** JWT (rol `admin`)

**Request Body:**

```json
{
  "username": "operador1",
  "password": "miPassword123",
  "role": "operator"
}
```

| Campo | Tipo | Requerido | Descripción |
|---|---|---|---|
| `username` | string | ✅ | Nombre de usuario (único) |
| `password` | string | ✅ | Contraseña (se guarda hasheada con bcrypt) |
| `role` | string | ❌ | `"admin"` o `"operator"` (default: `"operator"`) |

**Response `201 Created`:**

```json
{
  "message": "Usuario creado exitosamente",
  "user": {
    "id": 2,
    "username": "operador1",
    "role": "operator"
  }
}
```

**Response `403 Forbidden`:**

```json
{
  "error": "Solo administradores pueden crear usuarios"
}
```

**Response `409 Conflict`:**

```json
{
  "error": "El usuario ya existe"
}
```

---

### 🔑 `GET /api/auth/me`

Retorna los datos del usuario autenticado.

**Auth requerida:** JWT

**Response `200 OK`:**

```json
{
  "id": 1,
  "username": "admin",
  "role": "admin"
}
```

---

### 📋 `GET /api/drones`

Lista todos los drones registrados con su estado actual.

**Auth requerida:** JWT

**Response `200 OK`:**

```json
{
  "drones": [
    {
      "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
      "name": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
      "status": "idle",
      "lastHeartbeat": "2026-04-10T20:30:00Z",
      "lastLat": 10.4806,
      "lastLng": -66.9036,
      "lastAlt": 920.5,
      "isOnline": true,
      "createdAt": "2026-04-10T18:00:00Z",
      "lastTelemetry": {
        "orientation": { "pitch": 2.5, "roll": -1.3, "yaw": 185.7 },
        "gps": { "lat": 10.4806, "lng": -66.9036, "alt": 920.5 },
        "sensors": { "..." : "..." }
      },
      "currentMission": "MISSION-001"
    }
  ],
  "total": 1
}
```

| Campo | Tipo | Descripción |
|---|---|---|
| `droneId` | string | ID único del dron (`CHACOTUTO-UUID`) |
| `name` | string | Nombre del dron (por defecto = droneId) |
| `status` | string | Estado actual (ver tabla abajo) |
| `lastHeartbeat` | datetime | Último heartbeat recibido |
| `lastLat` | float | Última latitud GPS conocida |
| `lastLng` | float | Última longitud GPS conocida |
| `lastAlt` | float | Última altitud conocida (metros) |
| `isOnline` | bool | `true` si está conectado por WebSocket |
| `lastTelemetry` | object | Última telemetría completa (solo si online) |
| `currentMission` | string | ID de la misión activa (si tiene una) |

**Estados posibles del dron:**

| Status | Significado |
|---|---|
| `offline` | Desconectado del servidor |
| `idle` | Conectado, sin misión |
| `mission_received` | Misión recibida, esperando aceptación |
| `navigating` | Caminando al punto de inicio |
| `ready` | En punto de inicio, listo para misión |
| `in_mission` | Misión activa, transmitiendo telemetría |

---

### 📋 `GET /api/drones/:id`

Detalle de un dron específico, incluyendo sus misiones.

**Auth requerida:** JWT

**Parámetros URL:**

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | string | El `droneId` completo (ej: `CHACOTUTO-550e8400-...`) |

**Response `200 OK`:**

```json
{
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "name": "CHACOTUTO-550e8400-...",
  "status": "idle",
  "lastHeartbeat": "2026-04-10T20:30:00Z",
  "lastLat": 10.4806,
  "lastLng": -66.9036,
  "lastAlt": 920.5,
  "isOnline": true,
  "createdAt": "2026-04-10T18:00:00Z",
  "lastTelemetry": { "..." : "..." },
  "missions": [
    {
      "missionId": "MISSION-001",
      "droneId": "CHACOTUTO-550e8400-...",
      "status": "completed",
      "startLat": 10.4806,
      "startLng": -66.9036,
      "startAlt": 0.0,
      "createdAt": "2026-04-10T19:00:00Z",
      "updatedAt": "2026-04-10T19:15:00Z"
    }
  ]
}
```

**Response `404 Not Found`:**

```json
{
  "error": "Dron no encontrado"
}
```

---

### 📡 `GET /api/drones/:id/telemetry`

Historial de telemetría de un dron, con paginación.

**Auth requerida:** JWT

**Parámetros URL:**

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | string | El `droneId` |

**Query Parameters:**

| Parámetro | Tipo | Default | Descripción |
|---|---|---|---|
| `page` | int | `1` | Número de página |
| `limit` | int | `100` | Registros por página (máx 1000) |
| `mission` | string | — | Filtrar por `missionId` específico |

**Ejemplo:**

```
GET /api/drones/CHACOTUTO-550e8400-.../telemetry?page=1&limit=50&mission=MISSION-001
```

**Response `200 OK`:**

```json
{
  "droneId": "CHACOTUTO-550e8400-...",
  "page": 1,
  "limit": 50,
  "total": 342,
  "telemetry": [
    {
      "id": 1,
      "droneId": "CHACOTUTO-550e8400-...",
      "missionId": "MISSION-001",
      "timestamp": 1712783500123,
      "gyroX": 0.0123,
      "gyroY": -0.0456,
      "gyroZ": 0.0078,
      "accelX": 0.15,
      "accelY": 9.78,
      "accelZ": -0.32,
      "magX": 25.4,
      "magY": -12.8,
      "magZ": 42.1,
      "pitch": 2.5,
      "roll": -1.3,
      "yaw": 185.7,
      "lat": 10.4806,
      "lng": -66.9036,
      "alt": 920.5,
      "currentWaypoint": 2,
      "totalWaypoints": 5,
      "createdAt": "2026-04-10T19:05:00Z"
    }
  ]
}
```

**Campos de telemetría:**

| Campo | Unidad | Descripción |
|---|---|---|
| `gyroX/Y/Z` | rad/s | Velocidad angular (giroscopio) |
| `accelX/Y/Z` | m/s² | Aceleración lineal |
| `magX/Y/Z` | μT | Campo magnético |
| `pitch` | grados | Inclinación nariz (-180 a 180) |
| `roll` | grados | Inclinación lateral (-90 a 90) |
| `yaw` | grados | Rumbo / heading (0-360, 0=Norte) |
| `lat` | grados | Latitud GPS |
| `lng` | grados | Longitud GPS |
| `alt` | metros | Altitud sobre nivel del mar |
| `currentWaypoint` | int | Índice del waypoint actual |
| `totalWaypoints` | int | Total de waypoints de la misión |

> **Nota:** La telemetría se guarda a **1 registro por segundo** (muestreado). El dron envía a 10/s pero solo se persiste 1/s para no saturar la BD.

---

### 🎯 `GET /api/missions`

Lista todas las misiones.

**Auth requerida:** JWT

**Response `200 OK`:**

```json
{
  "missions": [
    {
      "missionId": "MISSION-a1b2c3d4",
      "droneId": "CHACOTUTO-550e8400-...",
      "status": "completed",
      "startLat": 10.4806,
      "startLng": -66.9036,
      "startAlt": 0.0,
      "waypoints": [
        { "id": 1, "missionId": "MISSION-a1b2c3d4", "lat": 10.4806, "lng": -66.9036, "alt": 50.0, "action": "takeoff", "index": 0 },
        { "id": 2, "missionId": "MISSION-a1b2c3d4", "lat": 10.4810, "lng": -66.9040, "alt": 100.0, "action": "waypoint", "index": 1 },
        { "id": 3, "missionId": "MISSION-a1b2c3d4", "lat": 10.4820, "lng": -66.9050, "alt": 50.0, "action": "land", "index": 2 }
      ],
      "createdAt": "2026-04-10T19:00:00Z",
      "updatedAt": "2026-04-10T19:15:00Z"
    }
  ],
  "total": 1
}
```

**Estados posibles de una misión:**

| Status | Significado |
|---|---|
| `pending` | Creada, enviada al dron, esperando respuesta |
| `accepted` | Dron aceptó la misión |
| `rejected` | Dron rechazó la misión |
| `in_progress` | Misión activa, dron volando |
| `completed` | Todos los waypoints completados |
| `cancelled` | Cancelada por el operador |

---

### 🎯 `GET /api/missions/:id`

Detalle de una misión específica.

**Auth requerida:** JWT

**Parámetros URL:**

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | string | El `missionId` (ej: `MISSION-a1b2c3d4`) |

**Response `200 OK`:**

```json
{
  "mission": {
    "missionId": "MISSION-a1b2c3d4",
    "droneId": "CHACOTUTO-550e8400-...",
    "status": "completed",
    "waypoints": [ "..." ],
    "startLat": 10.4806,
    "startLng": -66.9036,
    "startAlt": 0.0,
    "createdAt": "2026-04-10T19:00:00Z"
  },
  "telemetryCount": 342
}
```

---

### 🎯 `POST /api/missions`

Crea una misión y la envía al dron por WebSocket.

**Auth requerida:** JWT

**Request Body:**

```json
{
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "missionId": "MISSION-001",
  "waypoints": [
    {
      "lat": 10.4806,
      "lng": -66.9036,
      "alt": 50.0,
      "action": "takeoff",
      "index": 0
    },
    {
      "lat": 10.4810,
      "lng": -66.9040,
      "alt": 100.0,
      "action": "waypoint",
      "index": 1
    },
    {
      "lat": 10.4815,
      "lng": -66.9045,
      "alt": 100.0,
      "action": "waypoint",
      "index": 2
    },
    {
      "lat": 10.4820,
      "lng": -66.9050,
      "alt": 50.0,
      "action": "land",
      "index": 3
    }
  ],
  "startPoint": {
    "lat": 10.4806,
    "lng": -66.9036,
    "alt": 0.0,
    "action": "start",
    "index": 0
  }
}
```

| Campo | Tipo | Requerido | Descripción |
|---|---|---|---|
| `droneId` | string | ✅ | ID del dron destino (debe estar **online**) |
| `missionId` | string | ❌ | ID de la misión (se auto-genera si vacío) |
| `waypoints` | array | ✅ | Lista de waypoints (mínimo 1) |
| `startPoint` | object | ❌ | Punto de inicio (si null, usa primer waypoint) |

**Campos de cada waypoint:**

| Campo | Tipo | Descripción |
|---|---|---|
| `lat` | float | Latitud |
| `lng` | float | Longitud |
| `alt` | float | Altitud objetivo (metros) |
| `action` | string | `"takeoff"`, `"waypoint"`, `"land"`, `"start"` |
| `index` | int | Orden en la secuencia (0-based) |

**Response `201 Created`:**

```json
{
  "message": "Misión creada y enviada al dron",
  "mission": {
    "missionId": "MISSION-a1b2c3d4",
    "droneId": "CHACOTUTO-550e8400-...",
    "status": "pending"
  }
}
```

**Response `400 Bad Request`:**

```json
{
  "error": "El dron no está conectado"
}
```

> **¿Qué pasa en el fondo?**
> 1. Se guarda la misión y waypoints en la BD
> 2. Se envía `mission_assign` por WebSocket al dron
> 3. Se notifica al GCS por WebSocket

---

### 🎯 `DELETE /api/missions/:id`

Cancela una misión activa. Envía `mission_cancel` al dron por WebSocket.

**Auth requerida:** JWT

**Parámetros URL:**

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | string | El `missionId` |

**Response `200 OK`:**

```json
{
  "message": "Misión cancelada"
}
```

**Response `400 Bad Request`:**

```json
{
  "error": "La misión ya está completed"
}
```

---

## WebSocket: `/ws/drone`

Endpoint para conexiones de drones (teléfonos). **No requiere autenticación.**

**URL:** `ws://<HOST>:8080/ws/drone`

### Mensajes que el dron envía (App → Backend)

#### `register`
```json
{
  "type": "register",
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1712783400000
}
```
> Se envía **una vez** al conectar.

#### `heartbeat`
```json
{
  "type": "heartbeat",
  "droneId": "CHACOTUTO-550e8400-...",
  "status": "idle",
  "timestamp": 1712783402000
}
```
> Se envía **cada 2 segundos**.

#### `telemetry`
```json
{
  "type": "telemetry",
  "droneId": "CHACOTUTO-550e8400-...",
  "timestamp": 1712783500123,
  "sensors": {
    "gyroscope": { "x": 0.0123, "y": -0.0456, "z": 0.0078 },
    "accelerometer": { "x": 0.15, "y": 9.78, "z": -0.32 },
    "magnetometer": { "x": 25.4, "y": -12.8, "z": 42.1 }
  },
  "orientation": { "pitch": 2.5, "roll": -1.3, "yaw": 185.7 },
  "gps": { "lat": 10.4806, "lng": -66.9036, "alt": 920.5 },
  "mission": {
    "status": "in_progress",
    "currentWaypointIndex": 2,
    "totalWaypoints": 5
  }
}
```
> Se envía **cada 100ms** (10/s) solo durante misión activa.

#### `mission_ack`
```json
{
  "type": "mission_ack",
  "droneId": "CHACOTUTO-550e8400-...",
  "missionId": "MISSION-001",
  "status": "accepted",
  "timestamp": 1712783600000
}
```
> `status`: `"accepted"` o `"rejected"`

#### `mission_ready`
```json
{
  "type": "mission_ready",
  "droneId": "CHACOTUTO-550e8400-...",
  "missionId": "MISSION-001",
  "timestamp": 1712783700000
}
```

#### `mission_complete`
```json
{
  "type": "mission_complete",
  "droneId": "CHACOTUTO-550e8400-...",
  "missionId": "MISSION-001",
  "timestamp": 1712784000000
}
```

### Mensajes que el dron recibe (Backend → App)

#### `mission_assign`
```json
{
  "type": "mission_assign",
  "missionId": "MISSION-001",
  "waypoints": [
    { "lat": 10.4806, "lng": -66.9036, "alt": 50.0, "action": "takeoff", "index": 0 },
    { "lat": 10.4810, "lng": -66.9040, "alt": 100.0, "action": "waypoint", "index": 1 },
    { "lat": 10.4820, "lng": -66.9050, "alt": 50.0, "action": "land", "index": 2 }
  ],
  "startPoint": { "lat": 10.4806, "lng": -66.9036, "alt": 0.0, "action": "start", "index": 0 }
}
```

#### `mission_cancel`
```json
{
  "type": "mission_cancel",
  "missionId": "MISSION-001",
  "reason": "Cancelada por el operador"
}
```

---

## WebSocket: `/ws/gcs`

Endpoint para el dashboard GCS (interfaz web). **Requiere JWT en query parameter.**

**URL:** `ws://<HOST>:8080/ws/gcs?token=<JWT>`

### Mensaje al conectarse

Al conectarte recibes automáticamente el estado actual de todos los drones:

```json
{
  "type": "initial_state",
  "drones": [
    {
      "droneId": "CHACOTUTO-550e8400-...",
      "status": "idle",
      "isOnline": true,
      "lastHeartbeat": "2026-04-10T20:30:00Z"
    }
  ]
}
```

### Mensajes que recibes en tiempo real

#### `drone_online` — Un dron se conectó
```json
{
  "type": "drone_online",
  "droneId": "CHACOTUTO-550e8400-..."
}
```

#### `drone_offline` — Un dron se desconectó
```json
{
  "type": "drone_offline",
  "droneId": "CHACOTUTO-550e8400-...",
  "reason": "heartbeat_timeout"
}
```

#### `drone_status` — Cambio de estado del dron
```json
{
  "type": "drone_status",
  "droneId": "CHACOTUTO-550e8400-...",
  "status": "navigating"
}
```

#### `telemetry` — Telemetría en tiempo real (10/s por dron)
```json
{
  "type": "telemetry",
  "droneId": "CHACOTUTO-550e8400-...",
  "timestamp": 1712783500123,
  "sensors": { "..." },
  "orientation": { "..." },
  "gps": { "..." },
  "mission": { "..." }
}
```

#### `mission_ack` — Dron aceptó/rechazó misión
```json
{
  "type": "mission_ack",
  "droneId": "CHACOTUTO-550e8400-...",
  "missionId": "MISSION-001",
  "status": "accepted"
}
```

#### `mission_ready` — Dron listo para iniciar misión
```json
{
  "type": "mission_ready",
  "droneId": "CHACOTUTO-550e8400-...",
  "missionId": "MISSION-001"
}
```

#### `mission_complete` — Misión completada
```json
{
  "type": "mission_complete",
  "droneId": "CHACOTUTO-550e8400-...",
  "missionId": "MISSION-001"
}
```

#### `mission_assigned` — Una misión fue creada y enviada
```json
{
  "type": "mission_assigned",
  "missionId": "MISSION-001",
  "droneId": "CHACOTUTO-550e8400-...",
  "waypoints": [ "..." ],
  "timestamp": 1712784000000
}
```

#### `mission_cancelled` — Una misión fue cancelada
```json
{
  "type": "mission_cancelled",
  "missionId": "MISSION-001",
  "droneId": "CHACOTUTO-550e8400-..."
}
```

---

## Base de Datos

### Tabla `users`

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | uint (PK) | ID autoincremental |
| `username` | string (unique) | Nombre de usuario |
| `password` | string | Hash bcrypt (nunca se expone en API) |
| `role` | string | `"admin"` o `"operator"` |
| `created_at` | datetime | Fecha de creación |
| `updated_at` | datetime | Última actualización |

### Tabla `drones`

| Columna | Tipo | Descripción |
|---|---|---|
| `drone_id` | string (PK) | `CHACOTUTO-UUID` |
| `name` | string | Nombre del dron |
| `status` | string | Estado actual |
| `last_heartbeat` | datetime | Último heartbeat recibido |
| `last_lat` | float | Última latitud GPS |
| `last_lng` | float | Última longitud GPS |
| `last_alt` | float | Última altitud |
| `is_online` | bool | ¿Está conectado? |
| `created_at` | datetime | Primera vez que se registró |
| `updated_at` | datetime | Última actualización |

### Tabla `missions`

| Columna | Tipo | Descripción |
|---|---|---|
| `mission_id` | string (PK) | ID de la misión |
| `drone_id` | string (FK) | Dron asignado |
| `status` | string | Estado de la misión |
| `start_lat` | float | Latitud del punto de inicio |
| `start_lng` | float | Longitud del punto de inicio |
| `start_alt` | float | Altitud del punto de inicio |
| `created_at` | datetime | Fecha de creación |
| `updated_at` | datetime | Última actualización |

### Tabla `waypoints`

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | uint (PK) | ID autoincremental |
| `mission_id` | string (FK) | Misión a la que pertenece |
| `lat` | float | Latitud |
| `lng` | float | Longitud |
| `alt` | float | Altitud (metros) |
| `action` | string | `takeoff`, `waypoint`, `land`, `start` |
| `index` | int | Orden en la secuencia |

### Tabla `telemetry_logs`

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | uint (PK) | ID autoincremental |
| `drone_id` | string (FK) | Dron que envió la telemetría |
| `mission_id` | string (FK) | Misión activa al momento |
| `timestamp` | int64 | Unix ms del dron |
| `gyro_x/y/z` | float | Giroscopio (rad/s) |
| `accel_x/y/z` | float | Acelerómetro (m/s²) |
| `mag_x/y/z` | float | Magnetómetro (μT) |
| `pitch` | float | Inclinación nariz (grados) |
| `roll` | float | Inclinación lateral (grados) |
| `yaw` | float | Rumbo (grados) |
| `lat` | float | Latitud GPS |
| `lng` | float | Longitud GPS |
| `alt` | float | Altitud GPS (metros) |
| `current_waypoint` | int | Waypoint actual |
| `total_waypoints` | int | Total de waypoints |
| `created_at` | datetime | Timestamp del servidor |

> **Muestreo:** Se guarda **1 registro por segundo** por dron (no los 10/s que envía).

---

## Códigos de Error

| Código HTTP | Significado |
|---|---|
| `200` | OK — Petición exitosa |
| `201` | Created — Recurso creado (misión, usuario) |
| `400` | Bad Request — Datos inválidos o dron no conectado |
| `401` | Unauthorized — Token faltante o inválido |
| `403` | Forbidden — No tienes permisos (ej: no eres admin) |
| `404` | Not Found — Dron o misión no encontrado |
| `409` | Conflict — Recurso ya existe (ej: username duplicado) |
| `503` | Service Unavailable — No se pudo enviar al dron |

---

## Notas Importantes

1. **Timestamps** del dron son Unix milliseconds (`System.currentTimeMillis()`)
2. **GPS puede ser `null`** si el dispositivo no tiene señal
3. **Heartbeat timeout:** Si un dron no envía heartbeat por 10 segundos, se marca como `offline`
4. **Telemetría:** 10 msg/s al GCS en tiempo real, 1 msg/s guardado en BD
5. **El `droneId` es la clave primaria** — se genera en el teléfono al instalar la app
6. **JWT expira en 24 horas** — después hay que hacer login de nuevo
7. **`startPoint` puede ser `null`** al crear misión — se usa el primer waypoint
