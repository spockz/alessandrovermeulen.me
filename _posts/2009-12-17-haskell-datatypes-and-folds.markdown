--- 
layout: post
title: "Haskell Datatypes and Folds: Part I"
tags: 
- Haskell
- fold
comments: true
---
Welcome to this little explanation on how to determine the fold of a Haskell
datatype. First we'll look at how we define functions over lists, something
everyone starting with Haskell should be sufficiently familiar with, after which
we move on to the datatypes. You'll see different ways how to calculate the sum
of a list, how to fold over a list, what datatypes are and, how to fold over a
datatype, specifically the `BinTree a` datatype. Most importantly, I hope you
will grow to understand what a fold is, and why they are so important and useful
when programming Haskell.

During our functional programming course at Utrecht University I noticed
students having anxiety of datatypes and even more for folds. Not because they
couldn't grasp the workings of a single example but more a lack of a view on the
complete picture and lack of experience with the Haskell datatype way.

So, what are datatypes? Datatypes are a way of notating the abstract structure
of your data. There are several datastructures known to man, such as lists and
trees.
<h3>Lists</h3>
Lists in Haskell are used in several ways. Today we will look at how to
calculate the sum of a list. Intuitively you calculate the sum of a list by
adding its elements together. Starting with the first element and then
continuing on to the rest. This is how we literally translate that thought into
Haskell code:

``` haskell
sum :: [Int] -> Int
sum []     = 0
sum (x:xs) = x + sum xs
-- sum [1..4] = 10
```

Would we be calculating the product of the list, we'd do the same except we
multiply instead of adding.

``` haskell
product :: [Int] -> Int
product []     = 1
product (x:xs) = x * product xs
-- product [1..4] = 24
```

Looking at these two examples we can see that we have two similarities. We
always have recursion on the tail of the list and we do something with the head
of the list and the result of the recursion. In the first example we add them
together, and in product we multiply them. Another property of most functions
over lists is that there is a base case for the empty list. We call this the
<em>identity</em> of our function, sometimes also referred to as <em>unit</em>.
The identity of addition is 0 and the identity of multiplication is 1.

Now we look at one of the folds over lists defined in the prelude,
`foldr`. Make sure you know what each parameter stands for.
<strong>Note:</strong> There are some downsides to this function, mostly that it
will not work for large lists, you'll see more on this later.

``` haskell
foldr :: (a -> b -> b) -> b -> [a] -> b
foldr f z []     = z
foldr f z (x:xs) = f x (foldr f z xs)
-- The line above can also be written as:
-- foldr f z (x:xs) = x `f` (foldr f z xs)
```

Now, take a moment to let this function soak in and try to think of how you
would define `sum` and `product` in terms of this
`foldr`. Crucial at this point is to notice that we do not see any
type hardcoded in the type of `foldr` It may be helpful to look at
this fold, which is the identity function for lists:

``` haskell
idList :: [a] -> [a]
idList list = foldr (:) [] list
-- Recall: [1,2,3,4,5] == 1 : 2 : 3 : 4 : 5 : []
```


If we look at our sum function, the operator between each recursive call is
`(+)` and our base case is 0. So that's what we are going to use for
our sum in terms of foldr.

``` haskell
sumList :: [Int] -> Int
sumList list = foldr (+) 0 list
-- sumList [1..4] = 10
```

Thus far we only have lists of `Int` for our examples. However, for
sake of usability we will now move on to lists of numeric elements. Because
`(+)` is defined for all numbers. (If you wish to read more on this
subject please look up classes and instances.) Notice how the type of sumList
changes while it's definition remains unaltered.

``` haskell
sumList :: Num a => [a] -> a
sumList list = foldr (+) 0 list
```

Now that we have seen how we can determine the function and identity for our
fold from our recursive function to a definition in terms of `foldr`.
And most importantly, the foldr takes care of the recursive nature of the list
for us and the only thing it asks us in return for that is an operator and an
identity for your operation. So the fold can now be used to define several
operations on lists.

I told you about a problem of `foldr`. Depending on the size of your
list the above `sumList` may not work. Try the following on your
machine:

``` bash
$ ghci
Prelude> let sumList = foldr (+) 0
Prelude> sumList [1..1000000]
*** Exception: stack overflow
```

Now try:

``` bash
$ ghci
Prelude> let sumList = Data.List.foldl' (+) 0
Prelude> sumList [1..1000000]
500000500000
```

For a complete overview and analysis on `foldr` and
`foldl` please read <a
href="http://www.cs.nott.ac.uk/~gmh/fold.pdf">A tutorial on the universality and
expressiveness of fold</a>, by Graham Hutton.

<h3>Trees</h3>
First, let's take a look at how data structures in Haskell can be defined. In
short we have the `data` keyword, followed by zero or more `type
variables`, a `=` and then a number of `data
constructors` separated by a `|`.

Now that you have familiarized yourself with lists we can proceed to a slightly
more complicated datastructure. The Tree. In this example we will use a simple
binary tree. A binary tree can be denoted as follows:

``` haskell
data BinTree a = Node a (BinTree a) (BinTree a) | Leaf deriving (Show, Eq)
```

In this case `BinTree` is the <em>type constructor</em> and
`Node` and `Leaf` the data <em>data constructors</em>.

Remember that the goal of folds is to separate the implementation of the
recursion from the actual operation we want to execute on the datatype. So we
have one function, the <em>fold</em>, that takes care of the recursion and
several other functions that use this fold to specify certain semantics on the
datatype. We are going to calculate the sum of all elements in this tree.

The above binary tree has elements in the nodes and nothing in the leaves. You
can notice the recursive occurrence of ``BinTree a`' in the node. We
see two constructors in this datatype, `Node` and `Leaf`.
The Node constructor expects some value of type `a` and two subtrees
of type `BinTree a`, and the `Leaf` constructor has no
parameters. In Haskell:

``` haskell
Node :: a -> BinTree a -> BinTree a -> BinTree a
Leaf :: BinTree a
```

Notice that our list also has two constructors, namely the `(:)`
constructor that adds an element and the constructor for the empty list,
`[]`. Analogously we have the `Node` and `Leaf`
constructors. If we were allowed to use the `(:)` and `[]`
constructor in our own Haskell code, the datatype for a list would look like
this:

``` haskell
data List a = (:) a (List a) | []
```

Also, it is customary to keep the arguments for the functions in the same order
as the constructors are defined in the datatype, and to name the identifiers
containing the functions the same as the constructor function but in lowercase.
Applying this we get:

``` haskell
foldBinTree :: (a -> BinTree a -> BinTree a -> BinTree a) -> (BinTree a) -> BinTree a -> ??
foldBinTree node leaf = f
  where f (Node x left right) = node x (f left) (f right)
        f (Leaf)              = leaf
```

I left out the result type of this fold, try to find it yourself before
continuing.

By following our code we can see that every case, the case for `Node`
and the one for `Leaf`, result in a `BinTree a`.
Consequently the complete result of the function is a `BinTree a`

Now recall that we previously used a fold to calculate the sum of a list. With
that fold we were not restricted to a list. Which is also clearly visible by
looking at the type of `foldr` it contains only type variables. As we
want to calculate sum of the elements in this tree, which is something of type
`a` and not of type `BinTree a` we have to revise the type
of our fold. Notice the recurrence of the datatype in it's declaration,
`BinTree a`. We will replace all of these occurrences in the types of
our functions with a free type variable, say `r`:

``` haskell
foldBinTree :: (a -> r -> r -> r) -> (r) -> BinTree a -> r
foldBinTree node leaf = f
  where f (Node x left right) = node x (f left) (f right)
        f (Leaf)              = leaf
```

Keep in mind that I used 'eta reduction' in the above code. This means that the
last parameter (in this case the tree) isn't explicitly specified as it would
appear at the very end of the parameter list and ath the very end of the
definition. (<a href="http://www.haskell.org/haskellwiki/Eta_conversion">More on
eta reduction</a>.) So, now let's do something with this fold. Suppose we have
the following tree:

``` haskell
bintree = Node 1
            (Node 2
              (Node 3 Leaf Leaf)
              (Node 3 Leaf Leaf)
            )
            (Node 2
              (Node 3 Leaf Leaf)
              (Node 3 Leaf Leaf)
            )
```

A visual representation:
<pre>            1
          /  \
         2    2
        / \  / \
       3   3 3  3</pre>
Now we can define several traversals over this tree. Let's calculate the sum of
all values in the tree.

``` haskell
sumBinTree :: Num a => BinTree a -> a
sumBinTree = foldBinTree (\val res_left res_right -> val + res_left + res_right)
                         0
```
<h4>Algebras</h4>
The type of our foldBinTree is now more compact. However, there will be
datatypes that contain a larger sum of constructors that also may have more or
less parameters than in our case. Defining the type of the fold on those
datatypes as we have done before will inevitably lead to very clumsy type
signatures. The idea is that we split up the section that denotes our functions
into a separate type, namely the <em>algebra</em>.

So now we define an algebra for our data structure. To do this we again look at
each data constructor and determine it's type. The fold function takes care of
the recursion, so applying this thought consequently gives us the following
types and fold.

Keep in mind that we have to change the type of our fold, but not the definition
of the fold itself!

``` haskell
type BinTreeAlgebra a r = (a -> r -> r -> r, -- Node
                           r -- Leaf
                          )

foldBinTree :: BinTreeAlgebra a r -> BinTree a -> r
```

Now we can again define a sum on all `Num a` trees. Notice that we
went from to seperate parameters for the functions to one tuple with two
elements.

``` haskell
sumBinTree :: Num a => BinTree a -> a
sumBinTree = foldBinTree (\val res_left res_right -> val + res_left + res_right,
                          0)
```

``` bash
$ ghci bintree.hs
*Main> bintree
Node 1 (Node 2 (Node 3 Leaf Leaf) (Node 3 Leaf Leaf)) (Node 2 (Node 3 Leaf Leaf) (Node 3 Leaf Leaf))
*Main> sumBinTree bintree
19
```

That's it for now. If you did not understand everything by the end of this
article, don't panic. It will sink in eventually. Let this rest a day or two,
and then read this article again and you'll understand folds better. :)
