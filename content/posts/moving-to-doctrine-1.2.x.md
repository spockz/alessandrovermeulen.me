---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "Moving to Doctrine 1.2.x",
  "date": "2010-04-28T00:00:00+02:00",
  "slug": "moving-to-doctrine-1.2.x",
  "draft": false,
  "tags": [],
  "url": "/2010/04/28/moving-to-doctrine-1.2.x/",
  "aliases": [
    "/posts/moving-to-doctrine-1.2.x/"
  ]
}
---

Here are some updates on my previous article.
<a href="/2009/04/17/doctrine-meets-codeigniter/">Doctrine meets Codeigniter</a>.

Do not forget to add the following autoload directive to your hooks/doctrine.php and the doctrine.php cli:

```php
spl_autoload_register(array('Doctrine', 'modelsAutoload'));
```
