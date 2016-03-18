define(['app'], function(app) {
  app.registerService(
    'findById',
    function() {
      this.find = function(array, id) {
        for (var i=0; i<array.length; i++) {
          if (array[i].id == id) return array[i];
        }
      };
    }
  );
});
