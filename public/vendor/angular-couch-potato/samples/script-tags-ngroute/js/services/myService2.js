define([], function() {
  app.registerFactory(
    'myService2',
    [
      function() {
        var msg = 'This is from my service 2.';
        return {
          getMsg: function() {
            return msg;
          }
        };
      }
    ]
  );
});
