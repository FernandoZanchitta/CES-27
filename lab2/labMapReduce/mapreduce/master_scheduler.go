package mapreduce

import (
	"log"
	"sync"
)

// Schedules map operations on remote workers. This will run until InputFilePathChan
// is closed. If there is no worker available, it'll block.
func (master *Master) schedule(task *Task, proc string, filePathChan chan string) int {
	//////////////////////////////////
	// YOU WANT TO MODIFY THIS CODE //
	//////////////////////////////////
	master.retryOperationChan = make(chan *Operation, RETRY_OPERATION_BUFFER)
	master.completedOperations = 0
	master.totalOperations = 0
	var (
		wg, wgretry sync.WaitGroup
		filePath    string
		worker      *RemoteWorker
		operation   *Operation
		counter     int
		// ok          bool
	)

	log.Printf("Scheduling %v operations\n", proc)

	counter = 0
	for filePath = range filePathChan {
		operation = &Operation{proc, counter, filePath}
		counter++
		master.totalOperations++
		worker = <-master.idleWorkerChan
		wg.Add(1)

		go master.runOperation(worker, operation, &wg)
	}
	wg.Wait()

	for failedOperation := range master.retryOperationChan {
		worker = <-master.idleWorkerChan
		wgretry.Add(1)
		log.Printf("Retrying %v (ID: '%v' File: '%v' Worker: '%v')\n", failedOperation.proc, failedOperation.id, failedOperation.filePath, worker.id)
		go master.runOperation(worker, failedOperation, &wgretry)
		log.Printf("fim iteracao\n")
	}
	wgretry.Wait()
	// log.Printf("fim loop\n")
	// wgretry.Wait()
	// log.Printf("dps do wait\n")
	log.Printf("%vx %v operations completed\n", counter, proc)
	return counter
}

// runOperation start a single operation on a RemoteWorker and wait for it to return or fail.
func (master *Master) runOperation(remoteWorker *RemoteWorker, operation *Operation, wg *sync.WaitGroup) {
	//////////////////////////////////
	// YOU WANT TO MODIFY THIS CODE //
	//////////////////////////////////

	var (
		err  error
		args *RunArgs
	)

	log.Printf("Running %v (ID: '%v' File: '%v' Worker: '%v')\n", operation.proc, operation.id, operation.filePath, remoteWorker.id)

	args = &RunArgs{operation.id, operation.filePath}
	err = remoteWorker.callRemoteWorker(operation.proc, args, new(struct{}))

	if err != nil {
		log.Printf("Operation %v '%v' Failed. Error: %v\n", operation.proc, operation.id, err)
		master.retryOperationChan <- operation
		wg.Done()
		master.failedWorkerChan <- remoteWorker
	} else {
		//contabiliza o numero de operacoes completadas
		master.workersMutex.Lock()
		master.completedOperations++
		log.Printf("completedOperations: %v\n", master.completedOperations)
		master.workersMutex.Unlock()
		log.Printf("Operation %v '%v' Completed\n", operation.proc, operation.id)
		wg.Done()
		master.idleWorkerChan <- remoteWorker
		log.Printf("Worker '%v' is now idle\n", remoteWorker.id)

		if master.completedOperations == master.totalOperations {
			close(master.retryOperationChan)
		}
	}
}
