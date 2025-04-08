## Rules for MESSAGE BROKER CONTROLLER Implementation

### ✅ REQUIRED
- Use functions for controllers, not structs/interfaces
- Accept message client and usecases as parameters
- Extract data from topics and payloads
- Properly handle message acknowledgment
- Apply appropriate QoS levels
- Handle message deserialization errors gracefully
- Log all critical errors (to avoid silent failures)
- Follow naming convention: `[UsecaseName]SubscriberController`

### ❌ FORBIDDEN
- NEVER put business logic in controllers
- NEVER access databases or infrastructure directly (except through gateways)
- NEVER call more than one usecase from a single controller
- NEVER block message processing loops with synchronous operations
- NEVER panic on message processing errors (always recover)

### 💡 IMPORTANT TO REMEMBER
- Message broker controllers handle asynchronous communication
- Topic patterns may include wildcards that require parsing
- Consider message retention and delivery guarantees
- Error handling should be robust - failures shouldn't crash subscribers
- Message ordering may be important for some business processes
- Implement proper context timeout management
- Consider implementing dead-letter queues for failed messages
- Pay attention to concurrency and throughput considerations
- Keep subscriber controllers stateless when possible

### 📝 MQTT CONTROLLER EXAMPLE

```go
func DeviceHeartbeatSubscriberController(client mqtt.Client, u usecase.ProcessDeviceHeartbeat) {
    topic := "devices/+/heartbeat"
    qos := 1
    
    callback := func(client mqtt.Client, msg mqtt.Message) {
        // Extract device ID from topic
        parts := strings.Split(msg.Topic(), "/")
        if len(parts) != 3 {
            log.Printf("Invalid topic format: %s", msg.Topic())
            return
        }
        deviceID := parts[1]
        
        // Parse payload
        var heartbeatData usecase.ProcessDeviceHeartbeatReq
        if err := json.Unmarshal(msg.Payload(), &heartbeatData); err != nil {
            log.Printf("Error parsing heartbeat payload: %v", err)
            // Consider publishing to a dead-letter queue
            return
        }
        
        // Set device ID in request
        heartbeatData.DeviceID = deviceID
        
        // Create context with timeout
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        // Execute usecase
        _, err := u(ctx, heartbeatData)
        if err != nil {
            log.Printf("Error processing heartbeat: %v", err)
            // Consider retry logic or dead-letter queues here
            return
        }
    }
    
    token := client.Subscribe(topic, byte(qos), callback)
    if token.Wait() && token.Error() != nil {
        log.Printf("Error subscribing to %s: %v", topic, token.Error())
    } else {
        log.Printf("Subscribed to %s", topic)
    }
}
```

### 📝 RABBITMQ CONTROLLER EXAMPLE

```go
func OrderCreatedSubscriberController(ch *amqp.Channel, u usecase.ProcessNewOrder) {
    queue := "orders.created"
    exchange := "orders"
    routingKey := "created"
    
    // Declare exchange and queue
    err := ch.ExchangeDeclare(
        exchange,    // name
        "topic",     // type
        true,        // durable
        false,       // auto-deleted
        false,       // internal
        false,       // no-wait
        nil,         // arguments
    )
    if err != nil {
        log.Fatalf("Failed to declare exchange: %v", err)
    }
    
    q, err := ch.QueueDeclare(
        queue,       // name
        true,        // durable
        false,       // delete when unused
        false,       // exclusive
        false,       // no-wait
        nil,         // arguments
    )
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }
    
    err = ch.QueueBind(
        q.Name,      // queue name
        routingKey,  // routing key
        exchange,    // exchange
        false,       // no-wait
        nil,         // arguments
    )
    if err != nil {
        log.Fatalf("Failed to bind queue: %v", err)
    }
    
    msgs, err := ch.Consume(
        q.Name,      // queue
        "",          // consumer
        false,       // auto-ack
        false,       // exclusive
        false,       // no-local
        false,       // no-wait
        nil,         // args
    )
    if err != nil {
        log.Fatalf("Failed to register consumer: %v", err)
    }
    
    go func() {
        for msg := range msgs {
            var orderData usecase.ProcessNewOrderReq
            
            if err := json.Unmarshal(msg.Body, &orderData); err != nil {
                log.Printf("Error parsing order data: %v", err)
                // Reject message and don't requeue
                msg.Reject(false)
                continue
            }
            
            // Create context with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            
            // Execute usecase
            _, err := u(ctx, orderData)
            cancel() // Cancel context regardless of outcome
            
            if err != nil {
                log.Printf("Error processing order: %v", err)
                // Reject and requeue for retry
                msg.Reject(true)
                continue
            }
            
            // Acknowledge successful processing
            msg.Ack(false)
        }
    }()
    
    log.Printf("Listening for messages on %s", q.Name)
}
```

### 📝 KAFKA CONTROLLER EXAMPLE

```go
func UserEventSubscriberController(consumer *kafka.Consumer, u usecase.ProcessUserEvent) {
    topics := []string{"user-events"}
    
    err := consumer.SubscribeTopics(topics, nil)
    if err != nil {
        log.Fatalf("Failed to subscribe to topics: %v", err)
    }
    
    go func() {
        for {
            msg, err := consumer.ReadMessage(-1) // Wait indefinitely for message
            if err != nil {
                log.Printf("Error reading message: %v", err)
                continue
            }
            
            var eventData usecase.ProcessUserEventReq
            if err := json.Unmarshal(msg.Value, &eventData); err != nil {
                log.Printf("Error parsing user event: %v", err)
                // Handle bad message - possibly move to error topic
                continue
            }
            
            // Add metadata from Kafka message
            eventData.Timestamp = time.Unix(0, msg.Timestamp*int64(time.Millisecond))
            eventData.Topic = *msg.TopicPartition.Topic
            
            // Create context with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            
            // Execute usecase
            _, err = u(ctx, eventData)
            cancel() // Cancel context regardless of outcome
            
            if err != nil {
                log.Printf("Error processing user event: %v", err)
                // Consider retry logic or manual offset management
                continue
            }
            
            // Commit offset (auto-commit may also be enabled in config)
            _, err = consumer.CommitMessage(msg)
            if err != nil {
                log.Printf("Error committing offset: %v", err)
            }
        }
    }()
    
    log.Printf("Listening for user events on %v", topics)
}
```