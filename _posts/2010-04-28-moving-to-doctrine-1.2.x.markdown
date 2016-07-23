--- 
layout: post
title: Moving to Doctrine 1.2.x
tags: []
comments: true
---
Here are some updates on my previous article. <a href="http://alessandrovermeulen.me/2009/04/17/doctrine-meets-codeigniter/">Doctrine meets Codeigniter</a>.

Do not forget to add the following autoload directive to your hooks/doctrine.php and the doctrine.php cli:

[php]spl_autoload_register(array('Doctrine', 'modelsAutoload'));[/php]
