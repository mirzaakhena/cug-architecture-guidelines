## Rules for SCHEDULER CONTROLLER Implementation

### ✅ REQUIRED
- Use functions for controllers, not structs/interfaces
- Accept scheduler engine and usecases as parameters
- Create well-defined schedules (cron expressions or interval-based)
- Give each scheduled job a unique, descriptive identifier
- Implement proper error logging for scheduled job failures
- Follow naming convention: `[UsecaseName]SchedulerController`
- Include appropriate context timeout for each job
- Consider job execution metrics (success, failure, duration)

### ❌ FORBIDDEN
- NEVER put business logic in controllers
- NEVER access databases or infrastructure directly (except through gateways)
- NEVER call more than one usecase from a single controller
- NEVER hardcode schedule intervals or cron expressions (use configuration)
- NEVER implement overly long-running job functions (use timeouts)
- NEVER schedule jobs that could overlap with previous executions (unless explicitly designed for it)

### 💡 IMPORTANT TO REMEMBER
- Scheduler controllers are for time-triggered operations
- Cron expressions should be readable and well-documented
- Consider schedule distribution to avoid resource contention
- Be mindful of timezone issues in scheduling
- Implement appropriate retry mechanisms for transient failures
- Consider the implications of job execution delays
- Scheduled jobs should be idempotent when possible
- Include logging and monitoring for scheduled tasks
- Consider what happens if a scheduled job fails or takes too long
- Balance batch size and frequency for data processing jobs

### 📝 CRON-BASED SCHEDULER EXAMPLE

```go
func DailySummaryReportSchedulerController(scheduler *cronScheduler.Scheduler, u usecase.GenerateDailySummary) {
    // Schedule to run at 1 AM every day
    cronExpression := "0 1 * * *" // Minute, Hour, Day of Month, Month, Day of Week
    jobName := "daily-summary-report"
    
    // Register the job
    _, err := scheduler.AddJob(cronExpression, cron.Job{
        Name: jobName,
        Func: func() {
            // Add some jitter to prevent resource contention
            jitter := time.Duration(rand.Intn(300)) * time.Second
            time.Sleep(jitter)
            
            log.Printf("Starting scheduled job: %s", jobName)
            startTime := time.Now()
            
            // Create context with reasonable timeout
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
            defer cancel()
            
            // Create request with current date (previous day)
            yesterday := time.Now().AddDate(0, 0, -1)
            req := usecase.GenerateDailySummaryReq{
                Date: yesterday.Format("2006-01-02"),
            }
            
            // Execute usecase
            res, err := u(ctx, req)
            if err != nil {
                log.Printf("Error executing %s: %v", jobName, err)
                // Consider alerting for critical scheduled jobs
                return
            }
            
            duration := time.Since(startTime)
            log.Printf("Completed %s in %v - Generated %d reports", 
                jobName, duration, res.ReportCount)
        },
    })
    
    if err != nil {
        log.Fatalf("Failed to schedule job %s: %v", jobName, err)
    }
    
    log.Printf("Scheduled job %s with cron expression: %s", jobName, cronExpression)
}
```

### 📝 INTERVAL-BASED SCHEDULER EXAMPLE

```go
func HealthCheckSchedulerController(ticker *time.Ticker, u usecase.PerformSystemHealthCheck) {
    jobName := "system-health-check"
    
    go func() {
        for range ticker.C {
            log.Printf("Starting scheduled job: %s", jobName)
            startTime := time.Now()
            
            // Create context with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
            
            // Execute usecase
            res, err := u(ctx, usecase.PerformSystemHealthCheckReq{})
            cancel() // Cancel context regardless of outcome
            
            if err != nil {
                log.Printf("Health check failed: %v", err)
                // Consider implementing alerting for critical health checks
                continue
            }
            
            duration := time.Since(startTime)
            status := "HEALTHY"
            if !res.AllSystemsOperational {
                status = "DEGRADED"
            }
            
            log.Printf("Completed %s in %v - System status: %s", 
                jobName, duration, status)
        }
    }()
    
    log.Printf("Started interval-based job %s with period: %v", 
        jobName, ticker.Reset)
}
```

### 📝 DYNAMIC SCHEDULER EXAMPLE

```go
func DataSyncSchedulerController(scheduler *dynamicScheduler.Scheduler, u usecase.SyncExternalData) {
    jobName := "external-data-sync"
    
    // Function to register or update the job
    registerJob := func(syncInterval string) error {
        // Parse interval from configuration
        duration, err := time.ParseDuration(syncInterval)
        if err != nil {
            return fmt.Errorf("invalid sync interval format: %w", err)
        }
        
        // Define the job function
        jobFunc := func() {
            log.Printf("Starting scheduled job: %s", jobName)
            startTime := time.Now()
            
            // Create context with timeout proportional to interval (but capped)
            maxTimeout := 30 * time.Minute
            timeout := duration / 4
            if timeout > maxTimeout {
                timeout = maxTimeout
            }
            
            ctx, cancel := context.WithTimeout(context.Background(), timeout)
            defer cancel()
            
            // Create the request
            req := usecase.SyncExternalDataReq{
                FullSync: time.Now().Hour() == 2, // Full sync at 2 AM, delta otherwise
            }
            
            // Execute usecase
            res, err := u(ctx, req)
            if err != nil {
                log.Printf("Data sync failed: %v", err)
                return
            }
            
            duration := time.Since(startTime)
            log.Printf("Completed %s in %v - Synced %d records", 
                jobName, duration, res.RecordCount)
        }
        
        // Register or update the job with the scheduler
        return scheduler.RegisterOrUpdate(jobName, duration, jobFunc)
    }
    
    // Initial registration
    if err := registerJob("1h"); err != nil {
        log.Fatalf("Failed to register initial job %s: %v", jobName, err)
    }
    
    // Setup config watcher to update schedule based on configuration changes
    go func() {
        for configChange := range configWatcher.Changes() {
            if syncInterval, ok := configChange["data_sync_interval"]; ok {
                log.Printf("Updating %s schedule to: %s", jobName, syncInterval)
                if err := registerJob(syncInterval); err != nil {
                    log.Printf("Failed to update job %s: %v", jobName, err)
                }
            }
        }
    }()
    
    log.Printf("Registered dynamic job: %s", jobName)
}
```