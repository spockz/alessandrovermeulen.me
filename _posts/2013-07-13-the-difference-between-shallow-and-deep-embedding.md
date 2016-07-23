---
layout: post
title: "The difference between shallow and deep embedding"
date: 2013-07-13 23:06
comments: true
categories: Programming
tags:
  - LaTeX
  - TikZ
  - Graphviz
---

Deep and shallow embedding are terms associated with Domain Specific Languages
(DSL). A DSL is  a language geared toward a specific domain. The [dot language](http://www.graphviz.org/content/dot-language){:target="_blank"} is
an example of such a DSL for describing Graphs. Conceptually, a shallow
embedding captures the semantics of the data of the domain in a data type and
provides a *fixed* interpretation of the data, whereas a deep embedding goes
beyond this and captures the semantics of the operations on the domain enabling
*variable* interpretations.

We will illustrate this difference by embedding a simple expression language
with summation, multiplication and constants in
[Haskell](http://www.haskell.org). Haskell is especially well-suited for and
often used as a host language for embedded DSLs.

We express our language with the following interface. A type synonym `Exp` for
normal `Int`s and three separate functions representing summation,
multiplication, and constants.


``` haskell
type Exp = Int

plus  :: Exp -> Exp -> Exp
times :: Exp -> Exp -> Exp
const :: Int        -> Exp
```

We embedded the *data* of the domain in Haskell and provided functions for
construction of the  model and we can easily represent the calculation of an
expression as $4 + 6 * 8$ with the following lines of Haskell:

``` haskell
val = const 4 `plus` (const 6 `times` const 8)
```

The advantage of this embedding that calculating the value of our expression
is very fast. Other than the value we cannot determine anything else regarding
our expression. This becomes more problematic when we add variables to our
language.

We change our type to contain binding information and add two functions to
represent the assignment and usage of variables.

``` haskell
type Exp = ([String ⊨ Int], Int)

assign :: String -> Int -> Exp
var    :: String        -> Exp
```

And in our naivity we can write the expression $x + 6 * 8$ as follows:

``` haskell
val = var "x" `plus` (const 6 `times` const 8)
```

Obviously, evaluating this creates havoc! What is the value of `x`? We should,
of course, have introduced it first:

``` haskell
val = let "x" 4 (var "x" `plus` (const 6 `times` const 8))
```

Now we have assigned a value to `x` and we can safely use it in our
expression.

Had we used a deep embedding we could have prevented the cataclysmic error by
first checking whether each variable is assigned before it is used. We create
a deep embedding of our expression by using a Haskell data type.

``` haskell
data Exp where
  Plus   :: Exp -> Exp    -> Exp -- plus
  Times  :: Exp -> Exp    -> Exp -- times
  Const  :: Int           -> Exp -- const
  Assign :: String -> Int -> Exp -- assign
  Var    :: String        -> Exp -- var
```

Note that we do not specify *how* the bindings should be stored, only that
such a thing exists. We now define a function that checks whether we use a
variable before it is defined.[^folds]

``` haskell
useBeforeDefine :: Exp -> Bool
useBeforeDefine e = f []
  where
  f :: [String] -> Exp -> Bool
  f (Plus  l r) env      = useBeforeDefine l env || useBeforeDefine r env
  f (Times l r) env      = useBeforeDefine l env || useBeforeDefine r env
  f (Const _)   _        = False
  f (Assign var _ e) env = useBeforeDefine e (var : env)
  f (Var var)        env = not (var `elem` env)
```

With the function above we can *check* whether an expression is well-formed.
With our deep embedding we can even define transformations of our expression;
e.g. differentiate with respect to a variable.

```
diff :: Exp -> String -> Exp
diff (Plus  l r) dx      = diff l dx `Plus` diff r dx
diff (Times l r) dx      = (diff l dx `Times` r) `Plus` (l `Times` diff r dx)
diff (Const _)   _       = Const 0
diff (Assign var x e) dx = Assign var x (diff e dx)
diff (Var var)        dx | var == dx = Const 1
                         | otherwise = Const 0
```

Deep embedding allows us to utilize the semantics of our model by defining
multiple interpretations of our DSL. The downside is that just calculating the
value of our expression has become slower due to the added overhead of the
constructors, whereas the shallow embedding can be evaluated by only using
`Int`s.

In short:

* **Shallow embedding** should be used when you only need a single interpretation or
  when you are in a hurry.
* **Deep embedding** should be used in all other cases.

More reading material on this subject:

* This [presentation](http://www.cse.chalmers.se/~josefs/DSLTutorial/tutorialSlides.html) by Josef Svenningsson.
* [Combining Deep and Shallow Embedding for EDSL](http://www.cse.chalmers.se/~josefs/publications/TFP12.pdf) (Josef Svenningsson and Emil Axelsson, 2012)
* [Certifying Machine Code Safety: Shallow versus Deep Embedding](https://www4.in.tum.de/~nipkow/pubs/tphols04.html) (Martin Wildmoser and Tobias Nipkow, 2004)
* [Deep versus Shallow embeddings in Coq](http://cstheory.stackexchange.com/questions/1370/shallow-versus-deep-embeddings)

[^folds]: Most often you should use [folds](/2009/12/17/haskell-datatypes-and-folds/) ([2](/2010/01/03/haskell-datatypes-and-folds-part-ii/)) instead of this direct recursion.
