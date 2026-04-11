# CHACOTUTO — Protocolo WebSocket Dron ↔ Backend

> Este documento describe el protocolo de comunicación entre la app móvil (simulador de dron) y el backend.
> El backend debe implementar un servidor WebSocket que maneje estos mensajes.

---

## Arquitectura General

```
┌─────────────────────────┐         WebSocket (JSON)         ┌─────────────────────────┐
│   📱 App Móvil          │ ◄──────────────────────────────► │   🖥️ Backend (Go)        │
│   (Simulador de Dron)   │                                  │                         │
│                         │                                  │   ┌─────────────────┐   │
│  • Sensores (gyro,      │   ──── telemetry ──────►         │   │ DroneRegistry    │   │
│    accel, mag)           │   ──── register ──────►         │   │ (drones online)  │   │
│  • GPS                  │   ──── heartbeat ─────►         │   └─────────────────┘   │
│  • ID: CHACOTUTO-UUID   │   ──── mission_ack ───►         │                         │
│                         │   ──── mission_ready ──►         │   ┌─────────────────┐   │
│                         │   ──── mission_complete►         │   │ Mission Manager  │   │
│                         │                                  │   └─────────────────┘   │
│                         │   ◄─── mission_assign ──         │           │             │
│                         │   ◄─── mission_cancel ──         │           ▼             │
│                         │   ◄─── command ─────────         │   ┌─────────────────┐   │
└─────────────────────────┘                                  │   │ GCS Broadcast    │──►  🎮 GCS
                                                             └─────────────────────────┘
```

---

## Conexión WebSocket

- **URL esperada por la app:** `ws://<IP_SERVIDOR>:<PUERTO>/ws`
- **Protocolo:** WebSocket estándar (RFC 6455)
- **Formato:** JSON (UTF-8) en cada mensaje
- **Ping/Pong:** La app envía pings cada 15 segundos
- **Reconexión:** La app se reconecta automáticamente con backoff exponencial (2s → 32s)

---

## Identificación del Dron

Cada dispositivo genera un **ID único permanente** al instalar la app:

```
CHACOTUTO-550e8400-e29b-41d4-a716-446655440000
```

**Formato:** `CHACOTUTO-` + UUID v4

- Se genera una sola vez y se guarda en SharedPreferences
- **Nunca cambia** (a menos que se desinstale y reinstale la app)
- Se envía en **todos** los mensajes como campo `droneId`

---

## Mensajes: App → Backend

### 1. `register` — Registro del dron al conectarse

**Se envía:** Inmediatamente al abrir la conexión WebSocket.

```json
{
  "type": "register",
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1712783400000
}
```

**El backend debe:**
- Guardar el `droneId` en un registro de drones conectados
- Asociar la conexión WebSocket con ese `droneId`
- Notificar al GCS que un nuevo dron está online

---

### 2. `heartbeat` — Latido periódico

**Se envía:** Cada **2 segundos** mientras la conexión está activa.

```json
{
  "type": "heartbeat",
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "status": "idle",
  "timestamp": 1712783402000
}
```

**Valores posibles de `status`:**

| Status | Significado |
|---|---|
| `"idle"` | Sin misión, esperando |
| `"mission_received"` | Misión recibida, aún no aceptada |
| `"navigating"` | Caminando/moviéndose al punto de inicio |
| `"ready"` | En el punto de inicio, listo para despegar |
| `"in_mission"` | Misión activa, transmitiendo telemetría |

**El backend debe:**
- Actualizar el `lastHeartbeat` timestamp del dron
- Implementar un **timeout de desconexión** (ej: si no hay heartbeat en 10s → marcar como offline)
- Reenviar el estado al GCS si cambia

---

### 3. `telemetry` — Datos de sensores en tiempo real

**Se envía:** Cada **100ms** (10 veces por segundo) **solo durante una misión activa**.

```json
{
  "type": "telemetry",
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1712783500123,
  "sensors": {
    "gyroscope": {
      "x": 0.0123,
      "y": -0.0456,
      "z": 0.0078
    },
    "accelerometer": {
      "x": 0.15,
      "y": 9.78,
      "z": -0.32
    },
    "magnetometer": {
      "x": 25.4,
      "y": -12.8,
      "z": 42.1
    }
  },
  "orientation": {
    "pitch": 2.5,
    "roll": -1.3,
    "yaw": 185.7
  },
  "gps": {
    "lat": 10.480600,
    "lng": -66.903600,
    "alt": 920.5
  },
  "mission": {
    "status": "in_progress",
    "currentWaypointIndex": 2,
    "totalWaypoints": 5
  }
}
```

**Detalle de cada campo:**

#### `sensors.gyroscope` — Velocidad angular
| Campo | Unidad | Descripción |
|---|---|---|
| `x` | rad/s | Rotación alrededor del eje X |
| `y` | rad/s | Rotación alrededor del eje Y |
| `z` | rad/s | Rotación alrededor del eje Z |

#### `sensors.accelerometer` — Aceleración lineal
| Campo | Unidad | Descripción |
|---|---|---|
| `x` | m/s² | Aceleración lateral |
| `y` | m/s² | Aceleración vertical (≈9.8 en reposo) |
| `z` | m/s² | Aceleración frontal/trasera |

#### `sensors.magnetometer` — Campo magnético
| Campo | Unidad | Descripción |
|---|---|---|
| `x` | μT | Componente X del campo magnético |
| `y` | μT | Componente Y del campo magnético |
| `z` | μT | Componente Z del campo magnético |

#### `orientation` — Orientación fusionada del dispositivo
| Campo | Unidad | Rango | Descripción |
|---|---|---|---|
| `pitch` | grados | -180 a 180 | Inclinación nariz arriba/abajo |
| `roll` | grados | -90 a 90 | Inclinación lateral izq/der |
| `yaw` | grados | 0 a 360 | Rumbo / heading (0 = Norte) |

#### `gps` — Posición GPS
| Campo | Unidad | Descripción |
|---|---|---|
| `lat` | grados decimales | Latitud |
| `lng` | grados decimales | Longitud |
| `alt` | metros | Altitud sobre nivel del mar |

> **Nota:** `gps` puede ser `null` si el GPS no tiene señal aún.

#### `mission` — Progreso de la misión
| Campo | Tipo | Descripción |
|---|---|---|
| `status` | string | `"in_progress"` durante vuelo |
| `currentWaypointIndex` | int | Índice del waypoint actual (0-based) |
| `totalWaypoints` | int | Total de waypoints en la misión |

> **Nota:** `mission` puede ser `null` si no hay misión activa.

**El backend debe:**
- Almacenar la última telemetría de cada dron
- Reenviar al GCS en tiempo real (broadcast)
- Opcionalmente: guardar historial para reproducción

---

### 4. `mission_ack` — Confirmación de misión recibida

**Se envía:** Cuando el usuario acepta la misión en la app.

```json
{
  "type": "mission_ack",
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "missionId": "MISSION-001",
  "status": "accepted",
  "timestamp": 1712783600000
}
```

**Valores posibles de `status`:** `"accepted"` o `"rejected"`

**El backend debe:**
- Actualizar el estado de la misión
- Notificar al GCS

---

### 5. `mission_ready` — Dron listo en punto de inicio

**Se envía:** Cuando el usuario llega al punto de inicio y presiona "INICIAR MISIÓN".

```json
{
  "type": "mission_ready",
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "missionId": "MISSION-001",
  "timestamp": 1712783700000
}
```

**El backend debe:**
- Marcar el dron como "en misión"
- Notificar al GCS que la misión comenzó
- A partir de este momento recibirá `telemetry` cada 100ms

---

### 6. `mission_complete` — Misión finalizada

**Se envía:** Cuando el dron ha pasado por todos los waypoints.

```json
{
  "type": "mission_complete",
  "droneId": "CHACOTUTO-550e8400-e29b-41d4-a716-446655440000",
  "missionId": "MISSION-001",
  "timestamp": 1712784000000
}
```

**El backend debe:**
- Marcar la misión como completada
- Cambiar el estado del dron a `"idle"`
- Notificar al GCS
- Dejar de esperar `telemetry` a 100ms

---

## Mensajes: Backend → App

### 1. `mission_assign` — Asignar misión al dron

**El backend envía esto cuando el GCS crea una misión para un dron específico.**

```json
{
  "type": "mission_assign",
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

**Campos del waypoint:**

| Campo | Tipo | Descripción |
|---|---|---|
| `lat` | double | Latitud del waypoint |
| `lng` | double | Longitud del waypoint |
| `alt` | double | Altitud objetivo en metros |
| `action` | string | `"takeoff"`, `"waypoint"`, `"land"` |
| `index` | int | Orden en la secuencia (0-based) |

**`startPoint`:** Es el punto donde el dron (teléfono) debe estar físicamente antes de iniciar la misión. La app mostrará un mapa guiando al usuario a ese punto. Puede ser `null` (en ese caso usa el primer waypoint).

**La app hará:**
1. Mostrar los waypoints en el mapa
2. Esperar que el usuario acepte o rechace
3. Si acepta → guiar al usuario al `startPoint`
4. Al llegar (< 15 metros) → habilitar botón "INICIAR MISIÓN"

---

### 2. `mission_cancel` — Cancelar misión

**El backend envía esto cuando el GCS cancela una misión asignada.**

```json
{
  "type": "mission_cancel",
  "missionId": "MISSION-001",
  "reason": "Cancelada por el operador"
}
```

**La app hará:**
- Detener el streaming de telemetría
- Volver al estado idle
- Mostrar mensaje de cancelación

---

### 3. `command` — Comando genérico (extensible)

**Para futuros comandos que el backend necesite enviar.**

```json
{
  "type": "command",
  "command": "recalibrate",
  "params": {
    "sensor": "magnetometer"
  }
}
```

> Este tipo no es manejado actualmente por la app pero está reservado para expansión futura.

---

## Flujo Completo (Secuencia)

```
TIEMPO    APP (Dron)                    BACKEND                         GCS
─────── ────────────────────────────── ──────────────────────────────── ──────────────
  0s     WebSocket connect ──────────► Acepta conexión
  0s     register {droneId} ─────────► Guarda en DroneRegistry ────────► "Dron online"
  2s     heartbeat {idle} ───────────► Actualiza timestamp ────────────► Status update
  4s     heartbeat {idle} ───────────►
  ...    (repite cada 2s)

                                                                        Operador crea misión
                                       ◄────────────────────────────── POST /api/missions
         ◄─── mission_assign ──────── Envía misión al dron
         (muestra waypoints en mapa)
  Xs     mission_ack {accepted} ─────► Marca misión aceptada ──────────► "Misión aceptada"
         heartbeat {navigating} ─────►                                 ► Status update
         (usuario camina al punto)
         heartbeat {ready} ──────────►                                 ► Status update
         (presiona INICIAR MISIÓN)
  Ys     mission_ready ─────────────► Marca en misión ─────────────────► "Misión iniciada"

         ┌─ LOOP CADA 100ms ─────────────────────────────────────────────────────────┐
         │ telemetry {sensors, gps} ─► Almacena datos ─────────────────► Telemetría   │
         │ telemetry {sensors, gps} ─► Almacena datos ─────────────────► en tiempo    │
         │ telemetry {sensors, gps} ─► Almacena datos ─────────────────► real         │
         └───────────────────────────────────────────────────────────────────────────┘

         (todos los waypoints completados)
  Zs     mission_complete ──────────► Marca completada ────────────────► "Misión OK"
         heartbeat {idle} ──────────►                                  ► Status update
```

---

## API REST recomendada para el Backend

Además del WebSocket, el backend debería exponer estos endpoints para el GCS:

### `GET /api/drones`
Lista de todos los drones registrados con su último estado.

```json
{
  "drones": [
    {
      "droneId": "CHACOTUTO-550e8400-...",
      "status": "idle",
      "lastHeartbeat": "2026-04-10T20:30:00Z",
      "isOnline": true,
      "lastTelemetry": {
        "orientation": { "pitch": 2.5, "roll": -1.3, "yaw": 185.7 },
        "gps": { "lat": 10.4806, "lng": -66.9036, "alt": 920.5 }
      }
    }
  ]
}
```

### `POST /api/missions`
El GCS crea y asigna una misión a un dron específico.

```json
{
  "droneId": "CHACOTUTO-550e8400-...",
  "missionId": "MISSION-001",
  "waypoints": [
    { "lat": 10.4806, "lng": -66.9036, "alt": 50.0, "action": "takeoff", "index": 0 },
    { "lat": 10.4810, "lng": -66.9040, "alt": 100.0, "action": "waypoint", "index": 1 },
    { "lat": 10.4820, "lng": -66.9050, "alt": 50.0, "action": "land", "index": 2 }
  ],
  "startPoint": { "lat": 10.4806, "lng": -66.9036, "alt": 0.0, "action": "start", "index": 0 }
}
```

**El backend debe:**
1. Guardar la misión
2. Buscar el dron por `droneId` en el registro
3. Enviar `mission_assign` por WebSocket al dron
4. Responder al GCS con confirmación

### `GET /api/missions/{missionId}`
Estado actual de una misión.

### `DELETE /api/missions/{missionId}`
Cancela una misión (envía `mission_cancel` al dron).

---

## Modelo de datos recomendado para el Backend (Go)

```go
// Dron conectado
type ConnectedDrone struct {
    ID              string          `json:"droneId"`
    Conn            *websocket.Conn `json:"-"`
    Status          string          `json:"status"`
    LastHeartbeat   time.Time       `json:"lastHeartbeat"`
    LastTelemetry   *TelemetryData  `json:"lastTelemetry,omitempty"`
    CurrentMission  *Mission        `json:"currentMission,omitempty"`
    IsOnline        bool            `json:"isOnline"`
}

// Registro de drones
type DroneRegistry struct {
    mu     sync.RWMutex
    drones map[string]*ConnectedDrone
}

// Telemetría
type TelemetryData struct {
    Sensors     SensorPayload      `json:"sensors"`
    Orientation OrientationPayload `json:"orientation"`
    GPS         *GPSPayload        `json:"gps,omitempty"`
    Mission     *MissionStatus     `json:"mission,omitempty"`
    Timestamp   int64              `json:"timestamp"`
}

type SensorPayload struct {
    Gyroscope     Vec3 `json:"gyroscope"`
    Accelerometer Vec3 `json:"accelerometer"`
    Magnetometer  Vec3 `json:"magnetometer"`
}

type Vec3 struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
    Z float64 `json:"z"`
}

type OrientationPayload struct {
    Pitch float64 `json:"pitch"`
    Roll  float64 `json:"roll"`
    Yaw   float64 `json:"yaw"`
}

type GPSPayload struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
    Alt float64 `json:"alt"`
}

type MissionStatus struct {
    Status              string `json:"status"`
    CurrentWaypointIndex int   `json:"currentWaypointIndex"`
    TotalWaypoints      int    `json:"totalWaypoints"`
}

// Misión
type Mission struct {
    MissionID  string     `json:"missionId"`
    DroneID    string     `json:"droneId"`
    Waypoints  []Waypoint `json:"waypoints"`
    StartPoint *Waypoint  `json:"startPoint,omitempty"`
    Status     string     `json:"status"` // pending, accepted, in_progress, completed, cancelled
}

type Waypoint struct {
    Lat    float64 `json:"lat"`
    Lng    float64 `json:"lng"`
    Alt    float64 `json:"alt"`
    Action string  `json:"action"`
    Index  int     `json:"index"`
}
```

---

## Handler WebSocket recomendado (Go)

```go
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil { return }
    defer conn.Close()

    var droneID string

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            // Dron desconectado
            if droneID != "" {
                registry.MarkOffline(droneID)
                broadcastToGCS(map[string]any{
                    "type": "drone_offline", "droneId": droneID,
                })
            }
            break
        }

        var msg map[string]any
        json.Unmarshal(message, &msg)

        msgType, _ := msg["type"].(string)

        switch msgType {
        case "register":
            droneID, _ = msg["droneId"].(string)
            registry.Register(droneID, conn)
            broadcastToGCS(map[string]any{
                "type":    "drone_online",
                "droneId": droneID,
            })

        case "heartbeat":
            status, _ := msg["status"].(string)
            registry.UpdateHeartbeat(droneID, status)
            broadcastToGCS(map[string]any{
                "type":    "drone_status",
                "droneId": droneID,
                "status":  status,
            })

        case "telemetry":
            registry.UpdateTelemetry(droneID, msg)
            broadcastToGCS(msg) // Reenviar completo al GCS

        case "mission_ack":
            missionID, _ := msg["missionId"].(string)
            ackStatus, _ := msg["status"].(string)
            registry.UpdateMissionStatus(droneID, missionID, ackStatus)
            broadcastToGCS(msg)

        case "mission_ready":
            missionID, _ := msg["missionId"].(string)
            registry.UpdateMissionStatus(droneID, missionID, "in_progress")
            broadcastToGCS(msg)

        case "mission_complete":
            missionID, _ := msg["missionId"].(string)
            registry.CompleteMission(droneID, missionID)
            broadcastToGCS(msg)
        }
    }
}
```

---

## Resumen de frecuencias

| Mensaje | Dirección | Frecuencia | Condición |
|---|---|---|---|
| `register` | App → Backend | Una vez | Al conectar |
| `heartbeat` | App → Backend | Cada 2s | Siempre que esté conectado |
| `telemetry` | App → Backend | Cada 100ms | Solo durante misión activa |
| `mission_assign` | Backend → App | Bajo demanda | Cuando GCS asigna misión |
| `mission_cancel` | Backend → App | Bajo demanda | Cuando GCS cancela misión |
| `mission_ack` | App → Backend | Una vez | Al aceptar/rechazar misión |
| `mission_ready` | App → Backend | Una vez | Al presionar INICIAR MISIÓN |
| `mission_complete` | App → Backend | Una vez | Al completar todos los waypoints |

---

## Notas importantes

1. **Todos los timestamps** son Unix milliseconds (`System.currentTimeMillis()` en Kotlin)
2. **Todos los mensajes** incluyen `droneId` excepto los que vienen del backend
3. **El GPS puede ser `null`** si el dispositivo no tiene señal — el backend debe manejar esto
4. **La telemetría a 100ms** genera ~10 mensajes/segundo por dron — planificar capacidad
5. **El `droneId` es la clave primaria** — úsalo para todo el tracking
6. **`startPoint` puede ser `null`** en `mission_assign` — si es null, la app usa el primer waypoint
7. **Waypoint `action`** puede ser: `"takeoff"`, `"waypoint"`, `"land"`, `"start"`
