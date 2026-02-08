--------------------------- MODULE Orchestration ---------------------------
(***************************************************************************)
(* TLA+ specification of the orchestration layer in FibGo.                 *)
(*                                                                         *)
(* Models the concurrent execution of multiple Fibonacci calculators as    *)
(* implemented in internal/orchestration/orchestrator.go.                   *)
(*                                                                         *)
(* Key behaviors modeled:                                                   *)
(*   - N calculators run concurrently via errgroup                         *)
(*   - Each calculator sends progress updates to a bounded channel         *)
(*   - A display goroutine consumes progress updates                       *)
(*   - Channel is closed after all calculators complete                    *)
(*   - Display drains remaining messages then terminates                   *)
(*                                                                         *)
(* Go code reference:                                                       *)
(*   progressChan := make(chan fibonacci.ProgressUpdate,                    *)
(*                        len(calculators)*ProgressBufferMultiplier)        *)
(*   go progressReporter.DisplayProgress(...)                              *)
(*   g.Wait()       // wait for all calculators                            *)
(*   close(progressChan)                                                   *)
(*   displayWg.Wait()                                                      *)
(***************************************************************************)

EXTENDS Integers, Sequences, FiniteSets, TLC

CONSTANTS
    NumCalculators,     \* Number of concurrent calculators (1..N)
    ChannelCapacity     \* Buffer size = NumCalculators * 50

ASSUME NumCalculators \in Nat \ {0}
ASSUME ChannelCapacity \in Nat \ {0}

VARIABLES
    calcState,      \* Function: 1..N -> {"idle", "running", "done"}
    progressChan,   \* Sequence of messages (bounded buffer)
    results,        \* Function: 1..N -> {"nil", "value"}
    displayState,   \* "running" | "draining" | "done"
    channelOpen     \* Boolean: whether the channel is open

vars == <<calcState, progressChan, results, displayState, channelOpen>>

Calculators == 1..NumCalculators

TypeInvariant ==
    /\ calcState \in [Calculators -> {"idle", "running", "done"}]
    /\ progressChan \in Seq(Calculators)
    /\ Len(progressChan) <= ChannelCapacity
    /\ results \in [Calculators -> {"nil", "value"}]
    /\ displayState \in {"running", "draining", "done"}
    /\ channelOpen \in BOOLEAN

-----------------------------------------------------------------------------
(* Initial State *)
(* All calculators idle, channel empty and open, display running *)

Init ==
    /\ calcState = [i \in Calculators |-> "idle"]
    /\ progressChan = <<>>
    /\ results = [i \in Calculators |-> "nil"]
    /\ displayState = "running"
    /\ channelOpen = TRUE

-----------------------------------------------------------------------------
(* Actions *)

(* A calculator transitions from idle to running.
   Models: g.Go(func() { ... }) in orchestrator.go *)
CalculatorStart(i) ==
    /\ calcState[i] = "idle"
    /\ channelOpen = TRUE
    /\ calcState' = [calcState EXCEPT ![i] = "running"]
    /\ UNCHANGED <<progressChan, results, displayState, channelOpen>>

(* A running calculator sends a progress update to the channel.
   Non-blocking: if the channel is full, the update is dropped.
   Models: calculator.Calculate(ctx, progressChan, ...) sending updates *)
CalculatorReport(i) ==
    /\ calcState[i] = "running"
    /\ channelOpen = TRUE
    /\ IF Len(progressChan) < ChannelCapacity
       THEN progressChan' = Append(progressChan, i)
       ELSE progressChan' = progressChan  \* Drop if full (non-blocking send)
    /\ UNCHANGED <<calcState, results, displayState, channelOpen>>

(* A running calculator completes and writes its result.
   Models: results[idx] = CalculationResult{...} in the goroutine *)
CalculatorComplete(i) ==
    /\ calcState[i] = "running"
    /\ calcState' = [calcState EXCEPT ![i] = "done"]
    /\ results' = [results EXCEPT ![i] = "value"]
    /\ UNCHANGED <<progressChan, displayState, channelOpen>>

(* The display goroutine consumes one message from the channel head.
   Models: progressReporter.DisplayProgress reading from progressChan *)
DisplayConsume ==
    /\ displayState = "running"
    /\ Len(progressChan) > 0
    /\ progressChan' = Tail(progressChan)
    /\ UNCHANGED <<calcState, results, displayState, channelOpen>>

(* All calculators are done: close the channel.
   Models: g.Wait() followed by close(progressChan) *)
AllDone ==
    /\ \A i \in Calculators : calcState[i] = "done"
    /\ channelOpen = TRUE
    /\ channelOpen' = FALSE
    /\ displayState' = "draining"
    /\ UNCHANGED <<calcState, progressChan, results>>

(* Display drains remaining messages after channel is closed.
   Models: the DisplayProgress goroutine consuming remaining messages
   from the closed channel via range loop *)
DisplayDrain ==
    /\ displayState = "draining"
    /\ channelOpen = FALSE
    /\ IF Len(progressChan) > 0
       THEN /\ progressChan' = Tail(progressChan)
            /\ UNCHANGED displayState
       ELSE /\ displayState' = "done"
            /\ UNCHANGED progressChan
    /\ UNCHANGED <<calcState, results, channelOpen>>

-----------------------------------------------------------------------------
(* Next-State Relation *)

Next ==
    \/ \E i \in Calculators : CalculatorStart(i)
    \/ \E i \in Calculators : CalculatorReport(i)
    \/ \E i \in Calculators : CalculatorComplete(i)
    \/ DisplayConsume
    \/ AllDone
    \/ DisplayDrain

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

-----------------------------------------------------------------------------
(* Safety Properties *)

(* The channel is only closed after ALL calculators are done.
   This corresponds to the Go code: g.Wait() happens before close(). *)
SafeClose ==
    channelOpen = FALSE => \A i \in Calculators : calcState[i] = "done"

(* Each result slot is written at most once (no race condition).
   Once a result is "value", it stays "value". *)
NoRaceOnResults ==
    [][\A i \in Calculators :
        results[i] = "value" => results'[i] = "value"]_vars

(* The channel never exceeds its capacity. *)
ChannelBounded ==
    Len(progressChan) <= ChannelCapacity

(* Results are only written by completed calculators. *)
ResultsOnlyFromDone ==
    \A i \in Calculators :
        results[i] = "value" => calcState[i] = "done"

(* Type safety is always maintained. *)
SafetyInvariant ==
    /\ TypeInvariant
    /\ SafeClose
    /\ ChannelBounded
    /\ ResultsOnlyFromDone

-----------------------------------------------------------------------------
(* Liveness Properties *)

(* Deadlock freedom: eventually all calculators finish and display terminates.
   This requires weak fairness on all actions. *)
DeadlockFreedom ==
    <>(\A i \in Calculators : calcState[i] = "done" /\ displayState = "done")

(* Termination: the system eventually reaches a terminal state. *)
Termination ==
    <>(displayState = "done")

(* Every calculator that starts eventually completes. *)
CalculatorProgress ==
    \A i \in Calculators :
        calcState[i] = "running" ~> calcState[i] = "done"

(* The channel is eventually drained. *)
ChannelEventuallyDrained ==
    channelOpen = FALSE ~> Len(progressChan) = 0

=============================================================================
