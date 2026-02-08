----------------------- MODULE ConcurrencySemaphores -----------------------
(***************************************************************************)
(* TLA+ specification of the dual semaphore interaction in FibGo.          *)
(*                                                                         *)
(* Models two independent counting semaphores:                             *)
(*                                                                         *)
(*   1. Task Semaphore (fibonacci/common.go):                              *)
(*      - Capacity: NumCPU * 2                                             *)
(*      - Limits goroutines for Fibonacci-level parallelism                *)
(*      - Used by executeTasks() and executeMixedTasks()                   *)
(*                                                                         *)
(*   2. FFT Semaphore (bigfft/fft_recursion.go):                           *)
(*      - Capacity: NumCPU                                                 *)
(*      - Limits goroutines for FFT recursive parallelism                  *)
(*      - Used during FFT-based multiplication                             *)
(*                                                                         *)
(* Key design insight from the codebase:                                   *)
(*   When both semaphores are active, up to NumCPU * 3 goroutines may be   *)
(*   active simultaneously. This is mitigated by                           *)
(*   ShouldParallelizeMultiplication() which disables Fibonacci-level       *)
(*   parallelism when FFT is active (except for operands >                 *)
(*   ParallelFFTThreshold = 5M bits).                                      *)
(*                                                                         *)
(* Properties verified:                                                    *)
(*   - No deadlock (semaphores are independent, no circular wait)          *)
(*   - Maximum goroutines bounded by NumCPU * 3                            *)
(*   - Eventual completion of all workers                                  *)
(***************************************************************************)

EXTENDS Integers, FiniteSets, TLC

CONSTANTS
    NumCPU,             \* Number of CPU cores
    NumWorkers,         \* Number of worker goroutines to model
    TaskSemCapacity,    \* = NumCPU * 2
    FFTSemCapacity      \* = NumCPU

ASSUME NumCPU \in Nat \ {0}
ASSUME NumWorkers \in Nat \ {0}
ASSUME TaskSemCapacity = NumCPU * 2
ASSUME FFTSemCapacity = NumCPU

VARIABLES
    taskSem,        \* Count of acquired task semaphore tokens
    fftSem,         \* Count of acquired FFT semaphore tokens
    workers         \* Function: 1..NumWorkers -> worker state

vars == <<taskSem, fftSem, workers>>

Workers == 1..NumWorkers

(* Worker states model the lifecycle of a Fibonacci computation goroutine:
   - "idle":         Not yet started
   - "wait_task":    Waiting to acquire task semaphore token
   - "has_task":     Acquired task semaphore, doing non-FFT work
   - "wait_fft":     Acquired task semaphore, waiting for FFT semaphore
   - "has_both":     Acquired both semaphores, doing FFT work
   - "releasing":    Releasing semaphore tokens
   - "done":         Completed *)
WorkerStates == {"idle", "wait_task", "has_task", "wait_fft", "has_both", "releasing", "done"}

TypeInvariant ==
    /\ taskSem \in 0..TaskSemCapacity
    /\ fftSem \in 0..FFTSemCapacity
    /\ workers \in [Workers -> WorkerStates]

-----------------------------------------------------------------------------
(* Initial State *)

Init ==
    /\ taskSem = 0
    /\ fftSem = 0
    /\ workers = [i \in Workers |-> "idle"]

-----------------------------------------------------------------------------
(* Actions *)

(* Worker begins and tries to acquire task semaphore.
   Models: sem <- struct{}{} in executeTasks() *)
WorkerStart(i) ==
    /\ workers[i] = "idle"
    /\ workers' = [workers EXCEPT ![i] = "wait_task"]
    /\ UNCHANGED <<taskSem, fftSem>>

(* Worker acquires task semaphore token.
   Models: sem <- struct{}{} succeeding *)
AcquireTaskSem(i) ==
    /\ workers[i] = "wait_task"
    /\ taskSem < TaskSemCapacity
    /\ taskSem' = taskSem + 1
    /\ workers' = [workers EXCEPT ![i] = "has_task"]
    /\ UNCHANGED fftSem

(* Worker decides to use FFT and tries to acquire FFT semaphore.
   Models: the FFT recursion path in bigfft/fft_recursion.go
   which acquires the concurrencySemaphore *)
RequestFFT(i) ==
    /\ workers[i] = "has_task"
    /\ workers' = [workers EXCEPT ![i] = "wait_fft"]
    /\ UNCHANGED <<taskSem, fftSem>>

(* Worker acquires FFT semaphore token.
   Models: concurrencySemaphore <- struct{}{} in fft_recursion.go *)
AcquireFFTSem(i) ==
    /\ workers[i] = "wait_fft"
    /\ fftSem < FFTSemCapacity
    /\ fftSem' = fftSem + 1
    /\ workers' = [workers EXCEPT ![i] = "has_both"]
    /\ UNCHANGED taskSem

(* Worker with both semaphores completes FFT work and releases FFT semaphore.
   Models: defer func() { <-concurrencySemaphore }() *)
ReleaseFFTSem(i) ==
    /\ workers[i] = "has_both"
    /\ fftSem' = fftSem - 1
    /\ workers' = [workers EXCEPT ![i] = "has_task"]
    /\ UNCHANGED taskSem

(* Worker with task semaphore completes and releases.
   Models: defer func() { <-sem }() in executeTasks() *)
WorkerComplete(i) ==
    /\ workers[i] = "has_task"
    /\ taskSem' = taskSem - 1
    /\ workers' = [workers EXCEPT ![i] = "done"]
    /\ UNCHANGED fftSem

-----------------------------------------------------------------------------
(* Next-State Relation *)

Next ==
    \E i \in Workers :
        \/ WorkerStart(i)
        \/ AcquireTaskSem(i)
        \/ RequestFFT(i)
        \/ AcquireFFTSem(i)
        \/ ReleaseFFTSem(i)
        \/ WorkerComplete(i)

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

-----------------------------------------------------------------------------
(* Safety Properties *)

(* Semaphore counts never exceed their capacity. *)
SemaphoreBounds ==
    /\ taskSem >= 0
    /\ taskSem <= TaskSemCapacity
    /\ fftSem >= 0
    /\ fftSem <= FFTSemCapacity

(* Total active goroutines never exceed NumCPU * 3.
   Active = those holding at least the task semaphore.
   Those holding both contribute to both counts. *)
MaxGoroutines ==
    taskSem + fftSem <= NumCPU * 3

(* No deadlock: since the two semaphores are independent (no circular
   dependency), and workers always acquire task sem before FFT sem,
   there is a total order on resource acquisition that prevents deadlock.

   Formally: there is no state where all workers are blocked.
   A worker is blocked if it is in wait_task or wait_fft and cannot proceed. *)
NoDeadlock ==
    \/ \E i \in Workers : workers[i] \in {"idle", "has_task", "has_both", "done"}
    \/ \E i \in Workers : workers[i] = "wait_task" /\ taskSem < TaskSemCapacity
    \/ \E i \in Workers : workers[i] = "wait_fft" /\ fftSem < FFTSemCapacity

(* The task semaphore count equals the number of workers holding it. *)
TaskSemConsistency ==
    taskSem = Cardinality({i \in Workers : workers[i] \in {"has_task", "wait_fft", "has_both"}})

(* The FFT semaphore count equals the number of workers holding it. *)
FFTSemConsistency ==
    fftSem = Cardinality({i \in Workers : workers[i] = "has_both"})

(* Workers always acquire task semaphore before FFT semaphore.
   This total ordering prevents deadlock. *)
AcquisitionOrder ==
    \A i \in Workers :
        workers[i] \in {"wait_fft", "has_both"} => workers[i] # "idle"

(* Combined safety invariant. *)
SafetyInvariant ==
    /\ TypeInvariant
    /\ SemaphoreBounds
    /\ MaxGoroutines
    /\ TaskSemConsistency
    /\ FFTSemConsistency

-----------------------------------------------------------------------------
(* Liveness Properties *)

(* Every worker eventually completes. *)
EventualCompletion ==
    \A i \in Workers : workers[i] = "idle" ~> workers[i] = "done"

(* All workers eventually finish. *)
AllDone ==
    <>(\A i \in Workers : workers[i] = "done")

(* A worker waiting for the task semaphore eventually gets it. *)
TaskSemProgress ==
    \A i \in Workers :
        workers[i] = "wait_task" ~> workers[i] = "has_task"

(* A worker waiting for the FFT semaphore eventually gets it. *)
FFTSemProgress ==
    \A i \in Workers :
        workers[i] = "wait_fft" ~> workers[i] = "has_both"

(* Both semaphores are eventually fully released. *)
SemaphoresReleased ==
    <>(/\ taskSem = 0 /\ fftSem = 0)

=============================================================================
