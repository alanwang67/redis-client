package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"

	// "os"
	// "strconv"
	// "strings"
)


func createConfig(servers []string, threads uint64) []*redis.Client {
	i := uint64(0)
	// each thread's server to connect to 
	clientConnections := make([]*redis.Client , threads)

	for i < threads {
		c := redis.NewClient(&redis.Options{
			Addr:	  servers[i % uint64(len(servers))],
			Password: "srg", 
			DB:		  0,  
			Protocol: 2,  
		})
		clientConnections[i] = c
		i = i + 1
	}

	return clientConnections 
}

func main() {    
	threads, _ := strconv.ParseUint(os.Args[1], 10, 64)
	time, _ := strconv.ParseUint(os.Args[2], 10, 64)
	workload, _ := strconv.ParseUint(os.Args[3], 10, 64)

	serverAddresses := []string{"192.168.2.29:6379", "192.168.2.30:6379", "192.168.2.31:6379"} 
	clientConnections := createConfig(serverAddresses, threads)

	off_set := 5
	lower_bound := time.Duration(off_set) * time.Second
	upper_bound := time.Duration(uint64(off_set)+config.Time) * time.Second

	var l sync.Mutex

	set := false
	avg_time := float64(0)
	total_latency := time.Duration(0 * time.Microsecond)
	ops := uint64(0)

	var wg sync.WaitGroup
	var barrier sync.WaitGroup

	wg.Add(len(threads))
	barrier.Add(len(threads))

	i := uint64(0)
	for i < threads { 
		go func(clientId uint64, workload uint64, clientConnection *redis.Client) {
			index := uint64(0)
			serverId := i % uint64(3)
			var start_time time.Time
			var end_time time.Time
			var operation_start uint64
			var operation_end uint64
			var operation uint64
			var temp time.Duration 
	
			r := rand.New(rand.NewPCG(1, 2))
			z := rand.NewZipf(r, 3, 10, 100)
			ctx := context.Background()

			barrier.Done()
			barrier.Wait()
			defer wg.Done()
	
			log_time := false
			initial_time := time.Now()
			latency := time.Duration(0)
	
			for {
				if uint64(rand.IntN(99)) < workload {
					operation = uint64(1)
				} else {
					operation = uint64(0)
				}
	
				if !log_time && time.Since(initial_time) > lower_bound {
					start_time = time.Now()
					operation_start = index
					log_time = true
				}
				
				if log_time && time.Since(start_time) > (upper_bound) {
					end_time = time.Now()
					operation_end = index
					break
				}
	
				v := str(z.Uint64())
				sent_time := time.Now()
				if operation == 0 {
					err := clientConnection.Set(ctx, v, "a", 0).Err()
					if err != nil {
						panic(err)
					}  
				} else if operation == 1 {
					val, err := clientConnection.Get(ctx, "foo").Result()
					if err != nil {
						panic(err)
					}
				}

				temp = (time.Since(sent_time))
				latency = latency + temp 
				index = index + 1
			}
	
			l.Lock()
			if !set {
				avg_time = (end_time.Sub(start_time).Seconds())
				set = true
			} else {
				avg_time = (avg_time + (end_time.Sub(start_time).Seconds())) / 2
			}
			
			ops += operation_end - operation_start
			total_latency = total_latency + latency
			l.Unlock()
			
			return nil
		}(i, workload, clientConnections[i])
	
		i += 1
	}
	
	wg.Wait()
	
	fmt.Println("threads", config.Threads)
	fmt.Println("total_operations:", int(ops), "ops")
	fmt.Println("average_time:", int(avg_time), "sec")
	fmt.Println("throughput:", int(float64(ops)/(avg_time)), "ops/sec")
	fmt.Println("latency:", int(float64(total_latency.Microseconds())/float64(ops)), "us")
	
	return nil
}