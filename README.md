# countdown
Solver for the numbers game for the gameshow countdown

There is the countdown.go which is the application and then a cnt_slv package
To use the command line:
On a target of n and a list of number [...]
go run countdown.go --target n  [...]
e.g.
go run countdown.go --target 75 10 5 3 25

You can also add the option --seeks which will force it to search for the shortest solution
There are other options but they are now depriciated

As for the package it is a bit of a mess and needs tidy
The place to start is work_2_to_1
This takes 2 input numbers and returns a list of all the numbers that can be legally generated from those 2 numbers
for longer lists we work out all the way we can combine all the numbers we have available into things for work_2_to_1 to process.
i.e. a+b+c==(a+b)+c so we only need a 2 to 1 worker

For each set of input numbers [a,b,c] we generate a list of sets of numbers that can be generated from them
e.g.: [[a+b+c], [a+(b-c)], [a-(b+c)], ..., [a+b],[a-b],[a+c],[a+c],[b+c],...,[a],[b],[c]]
This is what the work_n function is trying to do, work any number of itens with any other number of items

Finally at the top level we have the PermuteN which tries every permutation of numbers into the work_n function
to encourage different arrangements

We also have a number of routines that self check the structures produced for sanity.

The end result of a call to PermuteN should be a list of everything you could possibly do with those numbers.
Possibly with a lot of redundancy in there but that's a tradeoff between time it takes to find the redundancy and remove it
vs cost of calculating over it.

Every time a new number is generated we tell a central map about it. This map function collects solutions found from an input set.
The map points to the tree of numbers that formed it.

The format of a Number is that it has its value, the operation used to generate it, and the list of numbers that went into it.
The difficulty field is just a rough approximation of how difficult we think the solution would be to find.
e.g.
&cnt_slv.Number{
        &cnt_slv.Number{
            Val:  19,
            list: {
                &cnt_slv.Number{
                    Val:        9,
                    list:       nil,
                    operation:  "I",
                    difficulty: 0,
                },
                &cnt_slv.Number{
                    Val:        10,
                    list:       nil,
                    operation:  "I",
                    difficulty: 0,
                },
            },
            operation:  "+",
            difficulty: 1,
        },
    },

so there we can see that 19 was formed by adding 9+10 <- which were given as Initial numbers

So when 19 is added as a new number it is added to the map if there is not already an entry for 19 in there or if the 19 that is in there is more difficult. 
The main() doesn't care about the final solution list as it only cares about the maped solution. So we end up with very large data structures for the garbage collector to clean up
Through performance work it seems that allocating memory and garbage collection are the main limitations on this algorithm.

It's really hard to know when a number is not interesting in order to do our own structure reclamation:
consider --target 6 3 7 4
Solution: 6=(3+(7-4))
When the first 3 is added it has difficulty 0 so will never be replaced.
at some point later we will create a new number 3 from 7-4 but that doesn't mean that we can re-use the data structure that the new 3 exists in.
We could trace the tree of solutions and look for things that are nto referenced by the number map. 
Unfortunatly because the solution map contains every number possible the same number is referenced many times so Just because we see it unused in one tree, it could be
used in another tree as a depandant of a useful number. We could do a garbage collection like algo where we tag useful items - but why would we be faster at that than the main GC?

