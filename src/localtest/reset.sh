COUNTER=11100
for i in `seq 11000 $COUNTER`;
do
  lsof -t -i tcp:$i | xargs kill
done
