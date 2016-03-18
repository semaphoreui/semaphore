app.config(function ($stateProvider, $couchPotatoProvider) {
  $stateProvider.state('users', {
    url: '/users',
    templateUrl: '/view/abstract',
    abstract: true
  })
  .state('users.add', {
    url: '/add',
    pageTitle: 'Add User',
    templateUrl: "/public/html/user/add.html",
    controller: 'AddUserCtrl',
    resolve: {
      dummy: $couchPotatoProvider.resolve(['controllers/user/add'])
    }
  })
  .state('users.list', {
    url: '/all',
    pageTitle: 'Users',
    templateUrl: "/public/html/user/list.html",
    controller: 'UsersCtrl',
    resolve: {
      dummy: $couchPotatoProvider.resolve(['controllers/user/users'])
    }
  })

  .state('user', {
    abstract: true,
    url: '/user/:user_id',
    templateUrl: '/public/html/abstract.html',
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
    templateUrl: '/public/html/user/view.html',
    resolve: {
      dummy: $couchPotatoProvider.resolve(['controllers/user/user'])
    }
  })
});