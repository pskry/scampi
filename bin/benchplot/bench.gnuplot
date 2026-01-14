#
# bench.gnuplot
#
# Usage:
#   gnuplot bench.gnuplot
#
#
set datafile separator ","
set terminal svg size 1200,800
set output "bench.svg"

set title "Benchmark evolution (median ms/op)"
set xlabel "Time"
set ylabel "ms/op"

set xdata time
set timefmt "%Y-%m-%dT%H:%M"
set format x "%m-%d\n%H:%M"

set grid
set key outside right

benchmarks = system("cut -d, -f1 bench.csv | tail -n +2 | sort -u")

plot for [b in benchmarks] \
     "bench.csv" using 2:(strcol(1) eq b ? $3 : 1/0) \
     with linespoints lw 2 title b
