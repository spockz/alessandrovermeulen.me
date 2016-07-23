--- 
layout: post
title: "Datatypes and Folds: Part II"
tags: 
- Haskell
- fold
- type algebra
comments: true
---
<h3>Non-mutually recursive and Mutually recursive datatypes</h3>
Welcome back. In this post we will look at creating type algebra's and folds for more complicated data types. In essence, this exercise will not be any more difficult than the previous ones, provided that you stick with the steps. Up in this part, are the non-mutually recursive datatypes and mutually recursive datatypes.

First up, data structures that are recursive into other data structures. The folds for these data structures are no more complex than their `simpler' forms. In fact, you could treat them the same, but for clarity we will treat them as distinct. If you want to know more about the previous part, it might help your search if you look for the keyword <em>catamorphism</em>.

<h4>Non-mutually recursive datastructures</h4>
Look at the following data structure. It's <em>data constructors</em> refer both back to the data structure itself and to another data structure. The other data structure has no reference back to this structure. This kind of data structures are quite common.

[haskell]
data MaybeTree a = Node (Maybe a) (MaybeTree a) (MaybeTree a)
                   | Leaf (Maybe a)
-- This one is already defined in the prelude
data Maybe a = Just a | Nothing
[/haskell]

Recall that in our previous folds and algebras we choose to replace our self-recursive types with a free variable <code>r</code>. In this case we also have a reference to another data structure, namely <code>Maybe a</code>. We will denote this by using another free variable, lets say <code>m</code>.

[haskell]
type MaybeTreeAlgebra a m r = (  ( m -&gt; r -&gt; r -&gt; r -- Node
                                 , m -&gt; r           -- Leaf
                                 )
                              ,  ( a -&gt; m           -- Just 
                                 , m                -- Nothing
                                 )               
                              )
[/haskell]

Notice that we've split up the the functions for each data type in separate tuples. This is to make the algebra somewhat more readable. It has one disadvantage though which we'll see later on.

We obtained this algebra by methodically looking at the types of the constructor functions and replacing any recursive types by their free variable counterparts, <code>r</code> for <code>MaybeTree a</code> and <code>m</code> for <code>Maybe a</code>.

Now we are going to construct the fold function. We will be doing this in exactly the same way as we did with all previous folds.

<ol>
  <li>Determine which datatypes are used (previously there was only one); This step is very simple if you already have the algebra.</li>
  <li>for each of these datatypes exhaustively define a fold function;</li>
<li>look at the result.</li>
</ol>

Following these steps again gives us the fold on our datatype. Please convince yourself that this is the case by defining <code>foldMaybeTree</code> yourself. 

If we do it exactly as described above it will provide us with the following fold. Notice that we gave meaningful names to every replacement function stating the datatype they are intended for. Except for the top level function which we called <code>f</code>. This is because writing out the whole name for that function everywhere would be to much of a hassle and if you use <code>f</code> for this way it will be quite clear in time.

[haskell]
foldMaybeTree' :: MaybeTreeAlgebra a m r -&gt; MaybeTree a -&gt; r
foldMaybeTree' ((node, leaf), (just, nothing)) = f
  where f (Node x l r) = node (maybe x) (f l) (f r)
        f (Leaf x)     = leaf (maybe x)
        maybe (Just x)  = just x
        maybe (Nothing) = nothing
[/haskell]

Notice that we are using a datatype here that is already defined in the prelude. It would be wise to check whether someone didn't already provide a fold function for our little datatype <code>Maybe a</code>. Because that's we essentially did, we inlined the fold for <code>Maybe a</code> inside our fold for <code>MaybeTree a</code>. Obviously this isn't a problem if you are sure that your datatype will stay `contained in' in your datatype. 

So we see that it is prudent to check for already existing folds if we use already existing datatypes. And it turns out that there already is a fold for <code>Maybe a</code>, namely <code>maybe</code>.

[haskell]
maybe :: b -&gt; (a -&gt; b) -&gt; Maybe a -&gt; b
maybe n _ Nothing  = n
maybe _ f (Just x) = f x
[/haskell]

So lets refine our little procedure to determine the fold of a datatype:

<ol>
  <li>Determine which datatypes are used (previously there was only one); This step is very simple if you already have the algebra.</li>
  <li>for each of these datatypes exhaustively define a fold function or check whether such a fold function already exists;</li>
<li>look at the result.</li>
</ol>

If we now use this knowledge our fold function can be written as follows:

[haskell]
foldMaybeTree :: MaybeTreeAlgebra a m r -&gt; MaybeTree a -&gt; r
foldMaybeTree ((node, leaf), (just, nothing)) = f
  where f (Node x l r) = node (fmaybe x) (f l) (f r)
        f (Leaf x)     = leaf (fmaybe x)
        fmaybe = maybe nothing just
[/haskell]

We've seen how to use algebra's but we haven't seen a step by step procedure for creating one. So let's add that to our procedure for creating our own folds. In our order to create the algebras as we've seen them here, with tupled functions you can follow this procedure:

For each data constructor in your datatype: create a function that takes the same parameters as the constructor function and replace every occurrence of a reference to one of your datatypes with a unique type variable (two occurrences to the same type should get the same variable). These functions should be grouped in tuples by their datatype. Do not forget to name all the type variables in the left-hand side of the type definition.
