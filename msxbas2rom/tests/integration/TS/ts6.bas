' ------------------
' TINY SPRITE TEST
' 64 PATTERNS LIMIT
' ------------------

FILE "sprite3.spr"   ' resource 0: msx1 sprite set

10 SCREEN 2, 2, 0    ' screen mode 2 (msx1)
20 SPRITE LOAD 0     ' load resource 0 (msx1 sprite set)

30 PUT SPRITE 0,(100, 20)
31 PUT SPRITE 1,(100, 60)
32 PUT SPRITE 2,(100, 60)
33 PUT SPRITE 3,(100, 60)
34 PUT SPRITE 4,(100,100)
35 PUT SPRITE 5,(100,100)
36 PUT SPRITE 6,(100,100)
37 PUT SPRITE 7,(100,140)
38 PUT SPRITE 8,(100,140)
39 PUT SPRITE 9,(100,140)

40 A$ = INPUT$(1)

50 SCREEN 0          
51 END

