package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	
	"os"
	"strconv"
	"sync"
	"math/rand/v2"
	"time"
)


func createConfig(servers []string, threads uint64) [][]*redis.Client {
	i := uint64(0)
	// each thread's server to connect to 
	clientConnections := make([][]*redis.Client , threads)

	for i < threads {
		j := uint64(0)
		// each server will have access to any of the servers
		c := make([]*redis.Client, len(servers))
		for j < uint64(len(servers)) {
			c[j] = redis.NewClient(&redis.Options{
				Addr:	  servers[j],
				Password: "srg", 
				DB:		  0,  
				Protocol: 2,  
			})
			j += 1
		}
		clientConnections[i] = c
		i = i + 1
	}

	return clientConnections 
}

func main() {    
	threads, _ := strconv.ParseUint(os.Args[1], 10, 64)
	t, _ := strconv.ParseUint(os.Args[2], 10, 64)
	workload, _ := strconv.ParseUint(os.Args[3], 10, 64)
	random, _ := strconv.ParseBool(os.Args[4])

	serverAddresses := []string{"192.168.2.29:6379", "192.168.2.30:6379", "192.168.2.31:6379"} 
	clientConnections := createConfig(serverAddresses, threads)

	off_set := 5
	lower_bound := time.Duration(off_set) * time.Second
	upper_bound := time.Duration(uint64(off_set)+t) * time.Second

	var l sync.Mutex

	set := false
	avg_time := float64(0)
	total_latency := time.Duration(0 * time.Microsecond)
	ops := uint64(0)
	switchServer := uint64(1000)

	var wg sync.WaitGroup
	var barrier sync.WaitGroup

	wg.Add(int(threads))
	barrier.Add(int(threads))

	i := uint64(0)
	for i < threads { 
		go func(clientId uint64, workload uint64, clientConnection []*redis.Client) error {
			index := uint64(0)
			writeServerId := uint64(0)
			readServerId := uint64(0)
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

				if random {
					if operation == uint64(0) && (index%switchServer == 0) {
						readServerId = uint64(rand.IntN(3)) 
					} else if operation == uint64(1) {
						writeServerId = uint64(0)
					}
				} else {
					if operation == uint64(0) {
						readServerId = clientId % uint64(3) 
					} else if operation == uint64(1) {
						writeServerId = uint64(0)
					}
				}

				v := strconv.FormatUint(z.Uint64(), 10)
				sent_time := time.Now()
				if operation == 0 {
					_, err := clientConnection[readServerId].Get(ctx, v).Result()
					temp = (time.Since(sent_time))
					if err != nil {
						fmt.Println(err)
						panic(err)
					}
				} else if operation == 1 {
					err := clientConnection[writeServerId].Set(ctx, v, "a", 0).Err()
					temp = (time.Since(sent_time))
					if err != nil {
						fmt.Println(err)
						panic(err)
					}  
				}

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
	
	fmt.Println("threads:", threads)
	fmt.Println("total_operations:", int(ops), "ops")
	fmt.Println("average_time:", int(avg_time), "sec")
	fmt.Println("throughput:", int(float64(ops)/(avg_time)), "ops/sec")
	fmt.Println("latency:", int(float64(total_latency.Microseconds())/float64(ops)), "us")
}