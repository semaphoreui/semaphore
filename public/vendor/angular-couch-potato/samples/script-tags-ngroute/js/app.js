// define([], function () {
  'use strict';

  window.app = angular.module('app', ['ngRoute', 'scs.couch-potato']);

  // setup the registerXXX functions
  couchPotato.configureApp(app);


  // set up RequireJS to know where components live
  require.config({
    baseUrl: 'js'
  });


  app.config(['$couchPotatoProvider', '$routeProvider',
    function( $couchPotatoProvider, $routeProvider) {

      $routeProvider.when('/view1',
        // Instead of the raw route object, we wrap it in a
        // call to resolveDependenciesProperty.  We use the provider
        // because we cannot inject $couchPotato service at config time.
        //
        // When the route is invoked, Couch Potato will resolve the dependencies
        // as Angular resolves the promise that holds the raw route object.
        $couchPotatoProvider.resolve({
          templateUrl:'partials/partial1.html',
          controller: 'MyCtrl1',
          dependencies: [
            //lazy/services/version is a dependency of directives/appVersion
            'directives/appVersion',
            'controllers/myCtrl1'
          ]
        })
      );

      $routeProvider.when('/view2',
        $couchPotatoProvider.resolve({
          templateUrl:'partials/partial2.html',
          controller: 'MyCtrl2',
          dependencies: [
            'controllers/myCtrl2',
            'filters/interpolator'
            //lazy/services/myService2 is an indirect dependency
          ]
        })
      );

      $routeProvider.otherwise({redirectTo: '/view1'});
    }
  ]);

  app.run(['$couchPotato', function($couchPotato) {
    // assign app.lazy so the registerXXX functions work
    app.lazy = $couchPotato;
  }]);

//   return app;

// });
