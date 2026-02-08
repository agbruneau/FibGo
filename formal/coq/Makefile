COQC = coqc
COQDEP = coqdep

VFILES = FastDoublingCorrectness.v FermatArithmetic.v
VOFILES = $(VFILES:.v=.vo)

.PHONY: all clean

all: $(VOFILES)

%.vo: %.v
	$(COQC) -R . FibGo $<

depend:
	$(COQDEP) -R . FibGo $(VFILES) > .depend

clean:
	rm -f *.vo *.vok *.vos *.glob .*.aux .depend

-include .depend
