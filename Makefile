install:
	git clone https://github.com/notional-labs/nursery 
	rm -rf nursery/.git
	mv nursery/go.mod nursery/go.m 
	go install .
	rm -rf nursery


clean:
	rm -rf nursery