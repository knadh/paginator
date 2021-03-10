# paginatior

paginator provides a simple abstracting for handling pagination requests and offset/limit generation for HTTP requests. The most common usecase is arbitrary queries that need to be paginated with query params coming in from a UI, for instance, /things/all?page=2&per_page=5. paginator can parse and sanitize these values and provide offset and limit values that can be passed to the database query there by avoiding boilerplate code for basic pagination. In addition, it can also generate HTML-ready page number series (Google search style).

## Features
- 0 boilerplate for reading pagination params from HTTP queries
- Automatic offset-limit calculator for DB queries
- Automatic sliding-window HTML pagination generation

![image](https://user-images.githubusercontent.com/547147/62465979-d73f8400-b7ad-11e9-98a0-dece2aac5d57.png)

## Usage
```go
    // Initialize global paginator instance.
    pg := paginator.New(paginator.Default())

    // Get page query params from an HTTP request.
    // The params to be picked up are defined in options
    // set by .Default() above.
    p := pg.NewFromURL(req.URL.Query())

    // or, pass page params directly, page and per_page.
    p := pg.New(1, 20)

    // Query your database with p.Offset and p.Limit.
    // Once you get the total number of results back
    // from the database, do:
    p.SetTotal(totalFromDB)

    // Generate HTML page numbers in a template.
    p.HTML()
```

## Example
Check out the [Alar dictionary glossary](https://alar.ink/glossary/kannada/english/%E0%B2%85) to see paginator in action.

Licensed under the MIT license.
