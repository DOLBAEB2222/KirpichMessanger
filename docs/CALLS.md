# Voice & Video Calls Documentation

This document describes the WebRTC-based voice and video calling system implemented in Stage 4.

## Architecture Overview

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   User A    │◄────►│   Server    │◄────►│   User B    │
│  (Caller)   │      │  (Signaling)│      │ (Callee)    │
└──────┬──────┘      └──────┬──────┘      └──────┬──────┘
       │                    │                    │
       │  1. Initiate Call  │                    │
       │───────────────────►│                    │
       │                    │  2. Notify via WS  │
       │                    │───────────────────►│
       │                    │                    │
       │  3. Accept/Reject  │                    │
       │◄───────────────────│◄───────────────────│
       │                    │                    │
       │  4. WebRTC Offer   │                    │
       │────────────────────────────────────────►│
       │                    │                    │
       │  5. WebRTC Answer  │                    │
       │◄────────────────────────────────────────│
       │                    │                    │
       │  6. ICE Exchange   │                    │
       │◄───────────────────────────────────────►│
       │                    │                    │
       │  7. P2P Connection │                    │
       │◄═══════════════════════════════════════►│
       │                    │                    │
       │  8. Hangup         │                    │
       │◄───────────────────────────────────────►│
```

## WebRTC Flow

### 1. Call Initiation (REST API)

The caller initiates a call via REST API:

```bash
POST /api/v1/calls
{
  "chat_id": "uuid",
  "recipient_id": "uuid",
  "call_type": "voice" | "video"
}
```

The server:
1. Creates a call record with status "ringing"
2. Sends `call:initiate` WebSocket event to recipient

### 2. Call Response (REST API)

The recipient responds via REST API:

```bash
PATCH /api/v1/calls/:call_id
{
  "accept": true | false
}
```

The server:
1. Updates call status to "accepted" or "rejected"
2. Sends `call:accepted` or `call:rejected` event to caller

### 3. WebRTC Signaling (WebSocket)

After call acceptance, peers exchange WebRTC signaling data via WebSocket:

#### Offer (Caller → Callee)
```json
{
  "type": "call:webrtc:offer",
  "data": {
    "call_id": "uuid",
    "offer": {
      "type": "offer",
      "sdp": "v=0\r\no=- 12345..."
    }
  }
}
```

#### Answer (Callee → Caller)
```json
{
  "type": "call:webrtc:answer",
  "data": {
    "call_id": "uuid",
    "answer": {
      "type": "answer",
      "sdp": "v=0\r\no=- 67890..."
    }
  }
}
```

#### ICE Candidates (Bidirectional)
```json
{
  "type": "call:webrtc:candidate",
  "data": {
    "call_id": "uuid",
    "candidate": {
      "candidate": "candidate:1 1 UDP 2130706431...",
      "sdpMid": "0",
      "sdpMLineIndex": 0
    }
  }
}
```

### 4. Call End (REST or WebSocket)

Either party can end the call:

**REST API:**
```bash
DELETE /api/v1/calls/:call_id
```

**WebSocket:**
```json
{
  "type": "call:hangup",
  "data": {
    "call_id": "uuid"
  }
}
```

## ICE/STUN/TURN Configuration

For NAT traversal, we use ICE (Interactive Connectivity Establishment) with STUN and TURN servers.

### STUN (Session Traversal Utilities for NAT)

STUN servers help peers discover their public IP addresses and ports.

**Public STUN servers:**
- `stun:stun.l.google.com:19302`
- `stun:stun.example.com:3478`

### TURN (Traversal Using Relays around NAT)

TURN servers relay media traffic when direct P2P connection is not possible.

**Configuration:**
```json
{
  "iceServers": [
    {
      "urls": ["turn:turn.example.com:3478?transport=udp"],
      "username": "user",
      "credential": "password"
    }
  ]
}
```

### Getting ICE Servers

Clients should fetch ICE server configuration before initiating calls:

```bash
GET /api/v1/calls/ice-servers
```

**Response:**
```json
{
  "iceServers": [
    {
      "urls": ["turn:turn.example.com:3478?transport=udp"],
      "username": "user",
      "credential": "password"
    },
    {
      "urls": ["stun:stun.example.com:3478"]
    },
    {
      "urls": ["stun:stun.l.google.com:19302"]
    }
  ]
}
```

## Client-Side Implementation

### JavaScript Example

```javascript
class CallManager {
  constructor(wsConnection) {
    this.ws = wsConnection;
    this.pc = null;
    this.localStream = null;
    this.remoteStream = null;
  }

  async initiateCall(chatId, recipientId, callType) {
    // 1. Create call via REST API
    const response = await fetch('/api/v1/calls', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        chat_id: chatId,
        recipient_id: recipientId,
        call_type: callType
      })
    });
    
    const call = await response.json();
    
    // 2. Setup WebRTC
    await this.setupWebRTC(call.id, true);
    
    return call;
  }

  async setupWebRTC(callId, isInitiator) {
    // Get ICE servers
    const iceServers = await this.fetchICEServers();
    
    // Create RTCPeerConnection
    this.pc = new RTCPeerConnection({ iceServers });
    
    // Get local media
    this.localStream = await navigator.mediaDevices.getUserMedia({
      audio: true,
      video: this.callType === 'video'
    });
    
    // Add tracks to peer connection
    this.localStream.getTracks().forEach(track => {
      this.pc.addTrack(track, this.localStream);
    });
    
    // Handle remote stream
    this.pc.ontrack = (event) => {
      this.remoteStream = event.streams[0];
      // Attach to video element
      document.getElementById('remoteVideo').srcObject = this.remoteStream;
    };
    
    // Handle ICE candidates
    this.pc.onicecandidate = (event) => {
      if (event.candidate) {
        this.ws.send(JSON.stringify({
          type: 'call:webrtc:candidate',
          data: {
            call_id: callId,
            candidate: event.candidate
          }
        }));
      }
    };
    
    if (isInitiator) {
      // Create offer
      const offer = await this.pc.createOffer();
      await this.pc.setLocalDescription(offer);
      
      this.ws.send(JSON.stringify({
        type: 'call:webrtc:offer',
        data: {
          call_id: callId,
          offer: offer
        }
      }));
    }
  }

  async handleOffer(callId, offer) {
    await this.pc.setRemoteDescription(new RTCSessionDescription(offer));
    
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);
    
    this.ws.send(JSON.stringify({
      type: 'call:webrtc:answer',
      data: {
        call_id: callId,
        answer: answer
      }
    }));
  }

  async handleAnswer(answer) {
    await this.pc.setRemoteDescription(new RTCSessionDescription(answer));
  }

  async handleCandidate(candidate) {
    await this.pc.addIceCandidate(new RTCIceCandidate(candidate));
  }

  async hangup(callId) {
    this.ws.send(JSON.stringify({
      type: 'call:hangup',
      data: { call_id: callId }
    }));
    
    this.cleanup();
  }

  cleanup() {
    if (this.pc) {
      this.pc.close();
      this.pc = null;
    }
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => track.stop());
      this.localStream = null;
    }
  }
}
```

## Call States

```
┌─────────┐
│ RINGING │◄── Initial state when call is created
└────┬────┘
     │
     ├──────► ┌──────────┐
     │        │ ACCEPTED │◄── Callee accepted
     │        └────┬─────┘
     │             │
     │             │ P2P connection established
     │             │
     │             ▼
     │        ┌────────┐
     │        │ ACTIVE │◄── Media flowing
     │        └───┬────┘
     │            │
     │            │ Any party hangs up
     │            │
     │            ▼
     │        ┌────────┐
     │        │ ENDED  │
     │        └────────┘
     │
     └──────► ┌──────────┐
              │ REJECTED │◄── Callee rejected
              └──────────┘
              
     ┌──────► ┌────────┐
     │        │ MISSED │◄── Timeout before answer
     │        └────────┘
     │
     └──────► ┌────────┐
              │  BUSY  │◄── Callee already in call
              └────────┘
```

## Security Considerations

1. **Authentication**: All call operations require valid JWT token
2. **Authorization**: Only call participants can access call data
3. **Encryption**: Media is encrypted via DTLS-SRTP in WebRTC
4. **TURN credentials**: Temporary credentials should be used for TURN

## Error Handling

### Common Errors

| Error | HTTP Code | Description |
|-------|-----------|-------------|
| Invalid call_id | 400 | Call ID is not a valid UUID |
| Not a member | 403 | User is not a member of the chat |
| Call not found | 404 | Call doesn't exist |
| Active call exists | 409 | Another call is already in progress |
| Call expired | 410 | Call timed out |

### WebSocket Errors

```json
{
  "type": "error",
  "code": "CALL_NOT_FOUND",
  "message": "The requested call does not exist"
}
```

## Testing

### Unit Testing

Run call-related tests:
```bash
go test ./... -run Call -v
```

### Manual Testing with curl

1. **Initiate call:**
```bash
curl -X POST http://localhost:8080/api/v1/calls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "chat_id": "chat-uuid",
    "recipient_id": "recipient-uuid",
    "call_type": "voice"
  }'
```

2. **Accept call:**
```bash
curl -X PATCH http://localhost:8080/api/v1/calls/$CALL_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"accept": true}'
```

3. **Get ICE servers:**
```bash
curl http://localhost:8080/api/v1/calls/ice-servers \
  -H "Authorization: Bearer $TOKEN"
```

4. **End call:**
```bash
curl -X DELETE http://localhost:8080/api/v1/calls/$CALL_ID \
  -H "Authorization: Bearer $TOKEN"
```

## Performance Considerations

1. **Bandwidth**: Voice calls need ~64 kbps, video calls need ~1-2 Mbps
2. **Latency**: Keep latency < 150ms for good quality
3. **Jitter**: Use jitter buffers to smooth out network variations
4. **Packet loss**: < 5% packet loss is acceptable for voice

## Future Enhancements

- [ ] Group calls (3+ participants)
- [ ] Screen sharing
- [ ] Call recording
- [ ] Call quality metrics
- [ ] Noise suppression
- [ ] Echo cancellation
- [ ] Video effects/filters
