define([
  'app',
  'factories/user'
], function(app) {
  app.config(function($stateProvider, $couchPotatoProvider) {
    $stateProvider

    .state('users', {
      url: '/users',
      templateUrl: '/view/abstract',
      abstract: true
    })
    .state('users.add', {
      url: '/add',
      pageTitle: 'Add User',
      templateUrl: "/view/user/add",
      controller: 'AddUserCtrl',
      resolve: {
        dummy: $couchPotatoProvider.resolve(['controllers/user/add'])
      }
    })
    .state('users.list', {
      url: '/all',
      pageTitle: 'Users',
      templateUrl: "/view/user/list",
      controller: 'UsersCtrl',
      resolve: {
        dummy: $couchPotatoProvider.resolve(['controllers/user/users'])
      }
    })

    .state('user', {
      abstract: true,
      url: '/user/:user_id',
      templateUrl: '/view/abstract',
      controller: ['$scope', 'user', function ($scope, user) {
        $scope.user = user;
      }],
      resolve: {
        user: ['User', '$stateParams', '$q', '$state', function (User, $stateParams, $q, $state) {
          var deferred = $q.defer();

          var user = new User($stateParams.user_id)
          user.get()
          .success(function (data, status) {
            user.data = data;
            deferred.resolve(user);
          })
          .error(function (data, status) {
            if (status == 404) {
              $state.transitionTo('homepage');
              return deferred.reject();
            }
          });

          return deferred.promise;
        }]
      }
    })

    .state('user.view', {
      url: '/',
      controller: 'UserCtrl',
      templateUrl: '/view/user/view',
      resolve: {
        dummy: $couchPotatoProvider.resolve(['controllers/user/user'])
      }
    })
  })
})
