S ::= program() # here to specify global program parameters

####################################################################################################
# Just a more convenient generation start point name                                               #
####################################################################################################

program @
  globals = glob-var-set(randint(0, 10))
::= {
  "int main()"
  "{"
    "return 0;"
  "}"
}

####################################################################################################
# sets                                                                                             #
####################################################################################################

glob-var-set n     ::= set-gen(glob-var, (), n)

set-gen _ _ 0 ::= ()
set-gen genf genf_in n @
  l = set-gen(genf, genf_in, sub(n, 1)),
  i = set-uniq(genf(genf_in,), l, genf, genf_in)
::= i:l
set-uniq i l genf genf_in ? in(i, l) ::= set-uniq(genf(genf_in,), l, genf, genf_in)
set-uniq i _ _ _ ::= i


####################################################################################################
# basic stuff                                                                                 #
####################################################################################################

glob-var _ ::= "g_" int() | *1000 "g_" identifier()

int ::= str(randint(1,max-int()))
max-int ::= 100


identifier ::= letter() letters()
letters ::= *4 letter() letters() | *2 ""
letter
::= "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j"
  | "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t"
  | "u" | "v" | "w" | "x" | "y" | "z"